// Copyright 2020 ETH Zurich
// Copyright 2020 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file contains the main router processing loop.

package main

import (
	"sync"

	"github.com/scionproto/scion/go/border/brconf"
	"github.com/scionproto/scion/go/border/internal/metrics"
	"github.com/scionproto/scion/go/border/qosqueues"
	"github.com/scionproto/scion/go/border/rcmn"
	"github.com/scionproto/scion/go/border/rctrl"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/assert"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/ringbuf"
	_ "github.com/scionproto/scion/go/lib/scrypto" // Make sure math/rand is seeded
)

const processBufCnt = 128

const maxNotificationCount = 512

const configFileLocation = "/home/marc/go/src/github.com/scionproto/scion/go/border/sample-config.yaml"

const noWorker = 1
const workLength = 32

var droppedPackets = 0

var queueToUse = &qosqueues.ChannelPacketQueue{}

// Router struct
type Router struct {
	// Id is the SCION element ID, e.g. "br4-ff00:0:2f".
	Id string
	// confDir is the directory containing the configuration file.
	confDir string
	// freePkts is a ring-buffer of unused packets.
	freePkts *ringbuf.Ring
	// sRevInfoQ is a channel for handling SignedRevInfo payloads.
	sRevInfoQ chan rpkt.RawSRevCallbackArgs
	// pktErrorQ is a channel for handling packet errors
	pktErrorQ chan pktErrorArgs
	// setCtxMtx serializes modifications to the router context. Topology updates
	// can be caused by a SIGHUP reload.
	setCtxMtx sync.Mutex

	config              qosqueues.InternalRouterConfig
	legacyConfig        qosqueues.RouterConfig
	notifications       chan *qosqueues.NPkt
	schedulerSurplus    qosqueues.Surplus
	schedulerSurplusMtx sync.Mutex
	workerChannels      [](chan *qosqueues.QPkt)
	forwarder           func(rp *rpkt.RtrPkt)
}

// NewRouter returns a new router
func NewRouter(id, confDir string) (*Router, error) {
	r := &Router{Id: id, confDir: confDir}
	if err := r.setup(); err != nil {
		return nil, err
	}

	r.initQueueing(configFileLocation)

	return r, nil
}

func (r *Router) initQueueing(location string) {

	//TODO: Figure out the actual path where the other config files are loaded
	// r.loadConfigFile("/home/vagrant/go/src/github.com/joelfischerr/scion/go/border/sample-config.yaml")
	err := r.loadConfigFile(location)

	if err != nil {
		log.Error("Loading config file failed", "error", err)
		panic("Loading config file failed")
	}

	//log.Debug("We have queues: ", "numberOfQueues", len(r.config.Queues))
	//log.Debug("We have rules: ", "numberOfRules", len(r.config.Rules))

	r.notifications = make(chan *qosqueues.NPkt, maxNotificationCount)
	r.forwarder = r.forwardPacket

	go r.drrDequer()

	r.workerChannels = make([]chan *qosqueues.QPkt, min(noWorker, len(r.config.Queues)))

	for i := range r.workerChannels {
		r.workerChannels[i] = make(chan *qosqueues.QPkt, workLength)

		go worker(&r.workerChannels[i])
	}

	//log.Debug("Finish init queueing")
}

// Start sets up networking, and starts go routines for handling the main packet
// processing as well as various other router functions.
func (r *Router) Start() {
	go func() {
		defer log.HandlePanic()
		r.PacketError()
		log.Info("handled packet error")
	}()
	go func() {
		defer log.HandlePanic()
		rctrl.Control(r.sRevInfoQ, cfg.General.ReconnectToDispatcher)
	}()
	//MS: check what types of queues are initialized and only run the go functions for these queues
	// go func() {
	// 	defer log.HandlePanic()
	// 	r.bscNotify()
	// }()
	go func() {
		defer log.HandlePanic()
		r.stochNotify()
	}()
	// go func() {
	// 	defer log.HandlePanic()
	// 	r.hbhNotify()
	// }()
}

// ReloadConfig handles reloading the configuration when SIGHUP is received.
func (r *Router) ReloadConfig() error {
	var err error
	var config *brconf.BRConf
	if config, err = r.loadNewConfig(); err != nil {
		return common.NewBasicError("Unable to load config", err)
	}
	if err := r.setupCtxFromConfig(config); err != nil {
		return common.NewBasicError("Unable to set up new context", err)
	}
	if err = r.loadConfigFile(configFileLocation); err != nil {
		return common.NewBasicError("Unable to load QoS config", err)
	}
	return nil
}

func (r *Router) handleSock(s *rctx.Sock, stop, stopped chan struct{}) {
	defer log.HandlePanic()
	defer close(stopped)
	pkts := make(ringbuf.EntryList, processBufCnt)
	dst := s.Conn.LocalAddr()
	log.Debug("handleSock starting", "addr", dst)
	for {
		n, _ := s.Ring.Read(pkts, true)
		if n < 0 {
			log.Debug("handleSock stopping", "addr", dst)
			return
		}
		for i := 0; i < n; i++ {
			rp := pkts[i].(*rpkt.RtrPkt)
			r.processPacket(rp)
			// the packet might still be queued so we can't release it here.
			// it is released in forwardPacket
			// rp.Release()
			pkts[i] = nil
		}
	}
}

// processPacket is the heart of the router's packet handling. It delegates
// everything from parsing the incoming packet, to routing the outgoing packet.
func (r *Router) processPacket(rp *rpkt.RtrPkt) {
	if assert.On {
		assert.Must(rp.DirFrom != rcmn.DirUnset, "DirFrom must be set")
		assert.Must(rp.Ingress.Dst != nil, "Ingress.Dst must be set")
		assert.Must(rp.Ingress.Src != nil, "Ingress.Src must be set")
		assert.Must(rp.Ctx != nil, "Context must be set")
		if rp.DirFrom == rcmn.DirLocal {
			assert.Must(rp.Ingress.IfID == 0, "Ingress.IfID must not be set for DirFrom==DirLocal")
		} else {
			assert.Must(rp.Ingress.IfID > 0, "Ingress.IfID must be set for DirFrom==DirExternal")
		}
	}
	l := metrics.ProcessLabels{
		IntfIn:  metrics.IntfToLabel(rp.Ingress.IfID),
		IntfOut: metrics.Drop,
	}
	// Assign a pseudorandom ID to the packet, for correlating log entries.
	rp.Id = log.NewDebugID().String()
	rp.Logger = log.New("rpkt", rp.Id)
	// XXX(kormat): uncomment for debugging:
	//rp.Debug("processPacket", "raw", rp.Raw)
	//rp.Debug("new Packet processing", "rpkt", rp) //MS
	if err := rp.Parse(); err != nil {
		r.handlePktError(rp, err, "Error parsing packet")
		l.Result = metrics.ErrParse
		metrics.Process.Pkts(l).Inc()
		return
	}
	//log.Debug("packet parsed", "rpkt", rp) //MS
	// Validation looks for errors in the packet that didn't break basic
	// parsing.
	valid, err := rp.Validate()
	if err != nil {
		r.handlePktError(rp, err, "Error validating packet")
		l.Result = metrics.ErrValidate
		metrics.Process.Pkts(l).Inc()
		return
	}
	//log.Debug("packet validated", "rpkt", rp) //MS
	if !valid {
		rp.Error("Error validating packet, no specific error")
		l.Result = metrics.ErrValidate
		metrics.Process.Pkts(l).Inc()
		return
	}
	// Check if the packet needs to be processed locally, and if so register hooks for doing so.
	rp.NeedsLocalProcessing()
	//log.Debug("packet checked for local processing", "Id", rp.Id, "Hooks", rp.Hooks()) //MS
	// Parse the packet payload, if a previous step has registered a relevant hook for doing so.
	if _, err := rp.Payload(true); err != nil {
		// Any errors at this point are application-level, and hence not
		// calling handlePktError, as no SCMP errors will be sent.
		rp.Error("Error parsing payload", "err", err)
		l.Result = metrics.ErrParsePayload
		metrics.Process.Pkts(l).Inc()
		return
	}
	//log.Debug("packet payload checked if necessary", "Id", rp.Id, "Hooks", rp.Hooks()) //MS
	// Process the packet, if a previous step has registered a relevant hook for doing so.
	if err := rp.Process(); err != nil {
		r.handlePktError(rp, err, "Error processing packet")
		l.Result = metrics.ErrProcess
		metrics.Process.Pkts(l).Inc()
		return
	}
	//log.Debug("processed packet if necessary", "Id", rp.Id, "Hooks", rp.Hooks()) //MS
	// Forward the packet. Packets destined to self are forwarded to the local dispatcher.
	// if err := rp.Route(); err != nil {
	// 	r.handlePktError(rp, err, "Error routing packet")
	// 	l.Result = metrics.ErrRoute
	// 	metrics.Process.Pkts(l).Inc()
	// }

	//log.Debug("Should queue packet")
	r.queuePacket(rp)
	// r.forwardPacket(rp);
}

func (r *Router) dropPacket(rp *rpkt.RtrPkt) {
	defer rp.Release()
	droppedPackets = droppedPackets + 1
	//log.Debug("Packet about to be dropped", "Rpkt", rp.Id)
	//log.Debug("Dropped Packet", "dropped", droppedPackets)

}

func (r *Router) forwardPacket(rp *rpkt.RtrPkt) {

	defer rp.Release()

	// Forward the packet. Packets destined to self are forwarded to the local dispatcher.
	if err := rp.Route(); err != nil {
		r.handlePktError(rp, err, "Error routing packet")
		l := metrics.ProcessLabels{
			IntfIn:  metrics.IntfToLabel(rp.Ingress.IfID),
			IntfOut: metrics.Drop,
		}
		l.Result = metrics.ErrRoute
		metrics.Process.Pkts(l).Inc()
	}
	//log.Debug("packet forwarded", "Id", rp.Id, "Hooks", rp.Hooks()) //MS
}

func (r *Router) queuePacket(rp *rpkt.RtrPkt) {

	//log.Debug("preRouteStep")
	//log.Debug("We have rules: ", "len(Rules)", len(r.config.Rules))

	queueNo := qosqueues.GetQueueNumberWithHashFor(&r.config, rp)
	qp := qosqueues.QPkt{Rp: rp, QueueNo: queueNo}

	r.workerChannels[(queueNo % noWorker)] <- &qp

}

func worker(workChannel *chan *qosqueues.QPkt) {

	for {
		qp := <-*workChannel
		queueNo := qp.QueueNo

		//log.Debug("Queuenumber is ", "queuenumber", queueNo)
		//log.Debug("Queue length is ", "len(r.config.Queues)", len(r.config.Queues))

		putOnQueue(queueNo, qp)
	}

}

func putOnQueue(queueNo int, qp *qosqueues.QPkt) {
	polAct := r.config.Queues[queueNo].Police(qp, queueNo == 1)
	profAct := r.config.Queues[queueNo].CheckAction()

	act := qosqueues.ReturnAction(polAct, profAct)

	if act == qosqueues.PASS {
		r.config.Queues[queueNo].Enqueue(qp)
		r.sendNotification(qp) //MS: remove this again
	} else if act == qosqueues.NOTIFY {
		r.config.Queues[queueNo].Enqueue(qp)
		r.sendNotification(qp)
	} else if act == qosqueues.DROPNOTIFY {
		r.dropPacket(qp.Rp)
		r.sendNotification(qp)
	} else if act == qosqueues.DROP {
		r.dropPacket(qp.Rp)
	} else {
		// This should never happen
		r.config.Queues[queueNo].Enqueue(qp)
	}
}

func (r *Router) sendNotification(qp *qosqueues.QPkt) {

	np := qosqueues.NPkt{Rule: qosqueues.GetRuleWithHashFor(&r.config, qp.Rp), Qpkt: qp}
	//log.Debug("Send notification method in router")
	restriction := r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning().InfoContent
	log.Debug("restrictions on information content", "restriction", restriction)
	np.Qpkt.Rp.RefInc(1) //should avoid the packet being dropped before we can create the scmp notification

	select {
	case r.notifications <- &np:
	default:
	}
}
