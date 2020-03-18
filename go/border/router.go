// Copyright 2016 ETH Zurich
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
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/scionproto/scion/go/border/brconf"
	"github.com/scionproto/scion/go/border/internal/metrics"
	"github.com/scionproto/scion/go/border/rcmn"
	"github.com/scionproto/scion/go/border/rctrl"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/assert"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/ringbuf"
	_ "github.com/scionproto/scion/go/lib/scrypto" // Make sure math/rand is seeded
	"gopkg.in/yaml.v2"
)

const processBufCnt = 128

const maxNotificationCount = 512

const configFileLocation = "/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/sample-config.yaml"

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

	config              InternalRouterConfig
	notifications       chan *qPkt
	flag                chan int
	schedulerSurplus    surplus
	schedulerSurplusMtx sync.Mutex
}

type surplus struct {
	surplus  int
	payments []int
}

// RouterConfig is what I am loading from the config file
type RouterConfig struct {
	Queues []packetQueue `yaml:"Queues"`
	Rules  []classRule   `yaml:"Rules"`
}

// InternalRouterConfig is what I am loading from the config file
type InternalRouterConfig struct {
	Queues []packetQueue
	Rules  []internalClassRule
}

// NewRouter returns a new router
func NewRouter(id, confDir string) (*Router, error) {
	r := &Router{Id: id, confDir: confDir}
	if err := r.setup(); err != nil {
		return nil, err
	}

	r.initQueueing()

	return r, nil
}

func (r *Router) loadConfigFile(path string) error {

	var internalRules []internalClassRule

	var rc RouterConfig

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	log.Info("Current Path is", "path", dir)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Info("yamlFile.Get ", "error", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &rc)
	if err != nil {
		log.Error("Unmarshal: ", "error", err)
		return err
	}

	for _, rule := range rc.Rules {
		intRule, err := convClassRuleToInternal(rule)
		if err != nil {
			log.Error("Error reading config file", "error", err)
		}
		internalRules = append(internalRules, intRule)
	}

	r.config = InternalRouterConfig{Queues: rc.Queues, Rules: internalRules}

	return nil
}

func (r *Router) initQueueing() {

	//TODO: Figure out the actual path where the other config files are loaded
	// r.loadConfigFile("/home/vagrant/go/src/github.com/joelfischerr/scion/go/border/sample-config.yaml")
	err := r.loadConfigFile(configFileLocation)

	if err != nil {
		log.Error("Loading config file failed", "error", err)
		panic("Loading config file failed")
	}

	// Initialise other data structures

	for i := 0; i < len(r.config.Queues); i++ {
		r.config.Queues[i].mutex = &sync.Mutex{}
		r.config.Queues[i].length = 0
		r.config.Queues[i].tb = tokenBucket{
			MaxBandWidth: r.config.Queues[i].PoliceRate,
			tokens:       r.config.Queues[i].PoliceRate,
			lastRefill:   time.Now(),
			mutex:        &sync.Mutex{}}
	}

	log.Info("We have queues: ", "numberOfQueues", len(r.config.Queues))

	r.flag = make(chan int, len(r.config.Queues))
	r.notifications = make(chan *qPkt, maxNotificationCount)
	r.forwarder = r.forwardPacket

	go func() {
		r.drrDequer()
	}()
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
	go func() {
		r.dequeuer()
	}()
	if err := r.startDiscovery(); err != nil {
		fatal.Fatal(common.NewBasicError("Unable to start discovery", err))
	}
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
	// Forward the packet. Packets destined to self are forwarded to the local dispatcher.
	// if err := rp.Route(); err != nil {
	// 	r.handlePktError(rp, err, "Error routing packet")
	// 	l.Result = metrics.ErrRoute
	// 	metrics.Process.Pkts(l).Inc()
	// }

	r.queuePacket(rp)
	// r.forwardPacket(rp);
}

func (r *Router) dropPacket(rp *rpkt.RtrPkt) {
	defer rp.Release()
	droppedPackets = droppedPackets + 1
	log.Debug("Dropped Packet", "dropped", droppedPackets)

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
}

func (r *Router) queuePacket(rp *rpkt.RtrPkt) {

	log.Debug("preRouteStep")

	// Put packets destined for 1-ff00:0:110 on the slow queue
	// Put all other packets from br2 on a faster queue but still delayed
	// At the moment no queue is slow

	queueNo := getQueueNumberForInternal(rp, &r.config.Rules)
	qp := qPkt{rp: rp, queueNo: queueNo}

	log.Info("Queuenumber is ", "queuenumber", queueNo)
	log.Info("Queue length is ", "len(r.config.Queues)", len(r.config.Queues))

	polAct := r.config.Queues[queueNo].police(&qp, queueNo == 1)
	profAct := r.config.Queues[queueNo].checkAction()

	act := returnAction(polAct, profAct)

	// if queueNo == 1 {
	// 	panic("We have received a packet on queue 1 🥳")
	// }

	if act == PASS {
		r.config.Queues[queueNo].enqueue(&qp)
	} else if act == NOTIFY {
		r.config.Queues[queueNo].enqueue(&qp)
		qp.sendNotification()
	} else if act == DROPNOTIFY {
		r.dropPacket(qp.rp)
		qp.sendNotification()
	} else if act == DROP {
		r.dropPacket(qp.rp)
	} else {
		// This should never happen
		r.config.Queues[queueNo].enqueue(&qp)
	}

	// According to gobyexample all sends are blocking and this is the standard way to do non-blocking sends (https://gobyexample.com/non-blocking-channel-operations)
	select {
	case r.flag <- queueNo:
	default:
	}

}
