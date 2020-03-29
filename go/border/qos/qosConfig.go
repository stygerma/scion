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

package qos

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/qos/qosscheduler"
	"github.com/scionproto/scion/go/border/rpkt"

	"github.com/scionproto/scion/go/lib/log"
	"gopkg.in/yaml.v2"
)

const maxNotificationCount = 512

type QosConfiguration struct {
	worker workerConfiguration

	config         qosqueues.InternalRouterConfig
	schedul        qosscheduler.SchedulerInterface
	legacyConfig   qosqueues.RouterConfig
	notifications  chan *qosqueues.NPkt
	workerChannels [](chan *qosqueues.QPkt)
	Forwarder      func(rp *rpkt.RtrPkt)

	droppedPackets int
}

type workerConfiguration struct {
	noWorker   int
	workLength int
}

func (q *QosConfiguration) SendToWorker(i int, qpkt *qosqueues.QPkt) {
	log.Debug("Start sending to worker")
	q.workerChannels[i] <- qpkt
	log.Debug("Finished sending to worker")
}

func (q *QosConfiguration) GetWorkerChannels() *[](chan *qosqueues.QPkt) {
	return &q.workerChannels
}

func (q *QosConfiguration) GetQueues() *[]qosqueues.PacketQueueInterface {
	return &q.config.Queues
}

func (q *QosConfiguration) GetQueue(ind int) *qosqueues.PacketQueueInterface {
	return &q.config.Queues[ind]
}

func (q *QosConfiguration) GetConfig() *qosqueues.InternalRouterConfig {
	return &q.config
}

func (q *QosConfiguration) GetLegacyConfig() *qosqueues.RouterConfig {
	return &q.legacyConfig
}

func InitQueueing(location string, forwarder func(rp *rpkt.RtrPkt)) (QosConfiguration, error) {

	qConfig := QosConfiguration{}

	noWorkers := max(1, min(3, len(qConfig.config.Queues)))

	qConfig.worker = workerConfiguration{noWorkers, 64}

	var err error
	qConfig.legacyConfig, qConfig.config, err = loadConfigFile(location)

	if err != nil {
		log.Error("Loading config file failed", "error", err)
		panic("Loading config file failed")
	}

	log.Debug("We have queues: ", "numberOfQueues", len(qConfig.config.Queues))
	log.Debug("We have rules: ", "numberOfRules", len(qConfig.config.Rules))

	qConfig.notifications = make(chan *qosqueues.NPkt, maxNotificationCount)
	qConfig.Forwarder = forwarder

	qConfig.schedul = &qosscheduler.RoundRobinScheduler{}
	qConfig.schedul.Init(qConfig.config)

	go qConfig.schedul.Dequeuer(qConfig.config, qConfig.Forwarder)

	qConfig.workerChannels = make([]chan *qosqueues.QPkt, qConfig.worker.noWorker)

	for i := range qConfig.workerChannels {
		qConfig.workerChannels[i] = make(chan *qosqueues.QPkt, qConfig.worker.workLength)

		log.Debug("Start worker", "workerno", i)
		go worker(&qConfig, &qConfig.workerChannels[i])
	}

	log.Debug("Finish init queueing")

	return qConfig, nil
}

func (qosConfig *QosConfiguration) QueuePacket(rp *rpkt.RtrPkt) {

	log.Debug("preRouteStep")
	log.Debug("We have rules: ", "len(Rules)", len(qosConfig.GetConfig().Rules))

	queueNo := qosqueues.GetQueueNumberWithHashFor(qosConfig.GetConfig(), rp)
	qp := qosqueues.QPkt{Rp: rp, QueueNo: queueNo}

	log.Debug("Our packet is", "QPkt", qp)
	log.Debug("Number of workers", "qosConfig.worker.noWorker", qosConfig.worker.noWorker)
	log.Debug("Sending it to worker", "workerNo", queueNo%qosConfig.worker.noWorker)

	select {
	case *qosConfig.schedul.GetMessages() <- true:
		log.Debug("sent message")
	default:
		log.Debug("no message sent")
	}

	qosConfig.SendToWorker(queueNo%qosConfig.worker.noWorker, &qp)

	log.Debug("Finished QueuePacket")

}

func worker(qosConfig *QosConfiguration, workChannel *chan *qosqueues.QPkt) {

	log.Debug("Started worker")
	for {
		log.Debug("Worker Waiting for new packet")
		qp := <-*workChannel
		log.Debug("Worker Received new packet")
		queueNo := qp.QueueNo

		log.Debug("Queuenumber is", "queuenumber", queueNo)
		log.Debug("Queue length is", "len(r.config.Queues)", len(qosConfig.config.Queues))

		log.Debug("Worker calling putOnQueue", "queueNo", queueNo, "packet", qp)
		putOnQueue(qosConfig, queueNo, qp)
	}

}

func putOnQueue(qosConfig *QosConfiguration, queueNo int, qp *qosqueues.QPkt) {
	log.Debug("putOnQueue")
	polAct := qosConfig.config.Queues[queueNo].Police(qp)
	log.Debug("Got polAct")
	profAct := qosConfig.config.Queues[queueNo].CheckAction()
	log.Debug("Got profAct")

	act := qosqueues.ReturnAction(polAct, profAct)

	log.Debug("Action is", "act", act)

	switch act {
	case qosqueues.PASS:
		log.Debug("pass")
		qosConfig.config.Queues[queueNo].Enqueue(qp)
	case qosqueues.NOTIFY:
		log.Debug("Notify")
		qosConfig.config.Queues[queueNo].Enqueue(qp)
		qosConfig.SendNotification(qp)
	case qosqueues.DROPNOTIFY:
		log.Debug("DROPNOTIFY")
		qosConfig.dropPacket(qp.Rp)
		qosConfig.SendNotification(qp)
	case qosqueues.DROP:
		log.Debug("DROP")
		qosConfig.dropPacket(qp.Rp)
	default:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
	}
}

func (qosConfig *QosConfiguration) SendNotification(qp *qosqueues.QPkt) {

	np := qosqueues.NPkt{Rule: qosqueues.GetRuleWithHashFor(&qosConfig.config, qp.Rp), Qpkt: qp}

	select {
	case qosConfig.notifications <- &np:
	default:
	}
}

func (qosConfig *QosConfiguration) dropPacket(rp *rpkt.RtrPkt) {
	defer rp.Release()
	qosConfig.droppedPackets += 1
	log.Debug("Dropped Packet", "dropped", qosConfig.droppedPackets)

}

func loadConfigFile(path string) (qosqueues.RouterConfig, qosqueues.InternalRouterConfig, error) {

	var internalRules []qosqueues.InternalClassRule
	var internalQueues []qosqueues.PacketQueueInterface

	var rc qosqueues.RouterConfig

	// dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// log.Debug("Current Path is", "path", dir)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return qosqueues.RouterConfig{}, qosqueues.InternalRouterConfig{}, err
	}
	err = yaml.Unmarshal(yamlFile, &rc)
	if err != nil {
		return qosqueues.RouterConfig{}, qosqueues.InternalRouterConfig{}, err
	}

	for _, rule := range rc.Rules {
		intRule, err := qosqueues.ConvClassRuleToInternal(rule)
		if err != nil {
			log.Error("Error reading config file", "error", err)
		}
		internalRules = append(internalRules, intRule)
	}

	var intQue qosqueues.PacketQueue

	for _, extQue := range rc.Queues {

		muta := &sync.Mutex{}
		mutb := &sync.Mutex{}

		queueToUse := &qosqueues.PacketSliceQueue{}

		log.Debug("We have loaded rc.Queues", "rc.Queues", rc.Queues)
		log.Debug("We have gotten the queue", "externalQueue", extQue.CongWarning)
		log.Debug("We have gotten the queue", "externalQueue", extQue.Name)
		intQue = convertExternalToInteralQueue(extQue)
		log.Debug("We have gotten the queue", "queue", intQue.CongWarning)
		log.Debug("We have gotten the queue", "queue", intQue.Name)
		queueToUse.InitQueue(intQue, muta, mutb)
		log.Debug("We have gotten the queue", "channelPacketQueue", queueToUse.GetPacketQueue().CongWarning)
		log.Debug("We have gotten the queue", "channelPacketQueue", queueToUse.GetPacketQueue().Name)
		internalQueues = append(internalQueues, queueToUse)
	}

	log.Debug("Loop over queues")
	for _, iq := range internalQueues {

		log.Debug("We have gotten the queue", "queue", iq.GetPacketQueue().CongWarning)
		log.Debug("We have gotten the queue", "queue", iq.GetPacketQueue().Name)

	}

	// r.legacyConfig = rc
	// r.config = qosqueues.InternalRouterConfig{Queues: internalQueues, Rules: internalRules}

	return rc, qosqueues.InternalRouterConfig{Queues: internalQueues, Rules: internalRules}, nil
}

func convertExternalToInteralQueue(extQueue qosqueues.ExternalPacketQueue) qosqueues.PacketQueue {

	pq := qosqueues.PacketQueue{
		Name:         extQueue.Name,
		ID:           extQueue.ID,
		MinBandwidth: extQueue.MinBandwidth,
		MaxBandWidth: extQueue.MaxBandWidth,
		PoliceRate:   convStringToNumber(extQueue.PoliceRate),
		MaxLength:    extQueue.MaxLength,
		Priority:     extQueue.Priority,
		CongWarning:  extQueue.CongWarning,
		Profile:      extQueue.Profile,
	}

	return pq
}

func convStringToNumber(bandwidthstring string) int {
	prefixes := map[string]int{
		"h": 2,
		"k": 3,
		"M": 6,
		"G": 9,
		"T": 12,
		"P": 15,
		"E": 18,
		"Z": 21,
		"Y": 24,
	}

	var num, powpow int

	for ind, str := range bandwidthstring {
		if val, contains := prefixes[string(str)]; contains {
			powpow = val
			num, _ = strToInt(bandwidthstring[:ind])
			return int(float64(num) * math.Pow(10, float64(powpow)))
		}
	}

	val, _ := strToInt(bandwidthstring)

	return val
}

func strToInt(str string) (int, error) {
	nonFractionalPart := strings.Split(str, ".")
	return strconv.Atoi(nonFractionalPart[0])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
