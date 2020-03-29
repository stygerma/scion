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
	q.workerChannels[i] <- qpkt
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

	qConfig.worker = workerConfiguration{3, 32}

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

	go qConfig.schedul.Dequeuer(qConfig.config, qConfig.Forwarder)

	qConfig.workerChannels = make([]chan *qosqueues.QPkt, min(qConfig.worker.noWorker, len(qConfig.config.Queues)))

	for i := range qConfig.workerChannels {
		qConfig.workerChannels[i] = make(chan *qosqueues.QPkt, qConfig.worker.workLength)

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

	qosConfig.SendToWorker(queueNo%qosConfig.worker.noWorker, &qp)

}

func worker(qosConfig *QosConfiguration, workChannel *chan *qosqueues.QPkt) {

	for {
		qp := <-*workChannel
		queueNo := qp.QueueNo

		log.Debug("Queuenumber is ", "queuenumber", queueNo)
		log.Debug("Queue length is ", "len(r.config.Queues)", len(qosConfig.config.Queues))

		putOnQueue(qosConfig, queueNo, qp)
	}

}

func putOnQueue(qosConfig *QosConfiguration, queueNo int, qp *qosqueues.QPkt) {
	polAct := qosConfig.config.Queues[queueNo].Police(qp)
	profAct := qosConfig.config.Queues[queueNo].CheckAction()

	act := qosqueues.ReturnAction(polAct, profAct)

	switch act {
	case qosqueues.PASS:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
	case qosqueues.NOTIFY:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
		qosConfig.SendNotification(qp)
	case qosqueues.DROPNOTIFY:
		qosConfig.dropPacket(qp.Rp)
		qosConfig.SendNotification(qp)
	case qosqueues.DROP:
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

		queueToUse := &qosqueues.ChannelPacketQueue{}

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
