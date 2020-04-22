// Copyright 2020 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
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
	"github.com/scionproto/scion/go/border/qos"
	"github.com/scionproto/scion/go/border/rcmn"
	"github.com/scionproto/scion/go/border/rctrl"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/assert"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/fatal"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/ringbuf"
	_ "github.com/scionproto/scion/go/lib/scrypto" // Make sure math/rand is seeded
)

const processBufCnt = 128

// TODO: this path should be configure in br.toml
const configFileLocation = "/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/sample-config.yaml"

const noWorker = 1
const workLength = 32

var droppedPackets = 0

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
<<<<<<< 9fe6884a06075d4213175d5498fd46b140b5d03e
	// qosConfig holds all data structures and state required for the quality of service
	// subsystem in the router
	qosConfig qos.QosConfiguration
}

// routerConfig is what I am loading from the config file
type routerConfig struct {
	Queues           []qosqueues.PacketQueueInterface
	Rules            []classRule
	SourceRules      map[addr.IA][]*classRule
	DestinationRules map[addr.IA][]*classRule
=======

<<<<<<< 1a0e0cc3f5b8eb1869b70b2fb2579fb32a8b0fd0
	config              qosqueues.InternalRouterConfig
	legacyConfig        qosqueues.RouterConfig
	notifications       chan *qosqueues.NPkt
	schedulerSurplus    qosqueues.Surplus
	schedulerSurplusMtx sync.Mutex
	workerChannels      [](chan *qosqueues.QPkt)
	forwarder           func(rp *rpkt.RtrPkt)
>>>>>>> Move ClassRule to qosqueues
=======
	// TODO: Put this configuration somewhere else
<<<<<<< f03ea997fce1af649af243cb79390d70594c2605
	config         qosqueues.InternalRouterConfig
	schedul        qosscheduler.SchedulerInterface
	legacyConfig   qosqueues.RouterConfig
	notifications  chan *qosqueues.NPkt
	workerChannels [](chan *qosqueues.QPkt)
	forwarder      func(rp *rpkt.RtrPkt)
>>>>>>> Put scheduler into its own package, but now there is an import cycle.
=======
	// config         qosqueues.InternalRouterConfig
	// schedul        qosscheduler.SchedulerInterface
	// legacyConfig   qosqueues.RouterConfig
	// notifications  chan *qosqueues.NPkt
	// workerChannels [](chan *qosqueues.QPkt)
	// forwarder      func(rp *rpkt.RtrPkt)

	qosConfig qos.QosConfiguration
>>>>>>> Suggestion for new file structure
}

// NewRouter returns a new router
func NewRouter(id, confDir string) (*Router, error) {
	r := &Router{Id: id, confDir: confDir}
	if err := r.setup(); err != nil {
		return nil, err
	}

	//TODO: Figure out the actual path where the other config files are loaded --> this path should be configure in br.toml
	// r.loadConfigFile("/home/vagrant/go/src/github.com/joelfischerr/scion/go/border/sample-config.yaml")
<<<<<<< f03ea997fce1af649af243cb79390d70594c2605
	r.loadConfigFile("/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/sample-config.yaml")

	log.Debug("We have queues: ", "numberOfQueues", len(r.config.Queues))
	log.Debug("We have rules: ", "numberOfRules", len(r.config.Rules))

	r.notifications = make(chan *qosqueues.NPkt, maxNotificationCount)
	r.forwarder = r.forwardPacket
=======
	r.qosConfig, _ = qos.InitQueueing(configFileLocation, r.forwardPacket)
>>>>>>> Suggestion for new file structure

	return r, nil
}

// Start sets up networking, and starts go routines for handling the main packet
// processing as well as various other router functions.
func (r *Router) Start() {
	go func() {
		defer log.HandlePanic()
		r.PacketError()
	}()
	go func() {
		defer log.HandlePanic()
		rctrl.Control(r.sRevInfoQ, cfg.General.ReconnectToDispatcher)
	}()
	if err := r.startDiscovery(); err != nil {
		fatal.Fatal(common.NewBasicError("Unable to start discovery", err))
	}
}

// TODO: Do we want to we also want to reload the queue config
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
<<<<<<< f03ea997fce1af649af243cb79390d70594c2605
=======
	if r.qosConfig, err = qos.InitQueueing(configFileLocation, r.forwardPacket); err != nil {
		return common.NewBasicError("Unable to load QoS config", err)
	}
>>>>>>> Suggestion for new file structure
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
	if err := rp.Parse(); err != nil {
		r.handlePktError(rp, err, "Error parsing packet")
		l.Result = metrics.ErrParse
		metrics.Process.Pkts(l).Inc()
		return
	}
	// Validation looks for errors in the packet that didn't break basic
	// parsing.
	valid, err := rp.Validate()
	if err != nil {
		r.handlePktError(rp, err, "Error validating packet")
		l.Result = metrics.ErrValidate
		metrics.Process.Pkts(l).Inc()
		return
	}
	if !valid {
		rp.Error("Error validating packet, no specific error")
		l.Result = metrics.ErrValidate
		metrics.Process.Pkts(l).Inc()
		return
	}
	// Check if the packet needs to be processed locally, and if so register hooks for doing so.
	rp.NeedsLocalProcessing()
	// Parse the packet payload, if a previous step has registered a relevant hook for doing so.
	if _, err := rp.Payload(true); err != nil {
		// Any errors at this point are application-level, and hence not
		// calling handlePktError, as no SCMP errors will be sent.
		rp.Error("Error parsing payload", "err", err)
		l.Result = metrics.ErrParsePayload
		metrics.Process.Pkts(l).Inc()
		return
	}
	// Process the packet, if a previous step has registered a relevant hook for doing so.
	if err := rp.Process(); err != nil {
		r.handlePktError(rp, err, "Error processing packet")
		l.Result = metrics.ErrProcess
		metrics.Process.Pkts(l).Inc()
		return
	}
<<<<<<< f03ea997fce1af649af243cb79390d70594c2605

	r.qosConfig.QueuePacket(rp)
}

func (r *Router) dropPacket(rp *rpkt.RtrPkt) {
	defer rp.Release()
	droppedPackets = droppedPackets + 1
	log.Debug("Dropped Packet", "dropped", droppedPackets)

	// TODO: We probably want some metrics here

}

=======
	// Forward the packet. Packets destined to self are forwarded to the local dispatcher.
	// if err := rp.Route(); err != nil {
	// 	r.handlePktError(rp, err, "Error routing packet")
	// 	l.Result = metrics.ErrRoute
	// 	metrics.Process.Pkts(l).Inc()
	// }

	log.Debug("Should queue packet")
	r.qosConfig.QueuePacket(rp)
	// r.forwardPacket(rp);
}

>>>>>>> Suggestion for new file structure
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
}
