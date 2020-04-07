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
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/scionproto/scion/go/border/qos/qosconf"
	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/qos/qosscheduler"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

const maxNotificationCount = 1024

type QosConfiguration struct {
	worker workerConfiguration

	config         qosqueues.InternalRouterConfig
	schedul        qosscheduler.SchedulerInterface
	legacyConfig   qosconf.ExternalConfig
	notifications  chan *qosqueues.NPkt
	workerChannels [](chan *qosqueues.QPkt)
	Forwarder      func(rp *rpkt.RtrPkt)

	droppedPackets int
}

type workerConfiguration struct {
	noWorker   int
	workLength int
}

func (qosConfig *QosConfiguration) SendToWorker(i int, qpkt *qosqueues.QPkt) {
	qosConfig.workerChannels[i] <- qpkt
}

func (qosConfig *QosConfiguration) GetWorkerChannels() *[](chan *qosqueues.QPkt) {
	return &qosConfig.workerChannels
}

func (qosConfig *QosConfiguration) GetQueues() *[]qosqueues.PacketQueueInterface {
	return &qosConfig.config.Queues
}

func (qosConfig *QosConfiguration) GetQueue(ind int) *qosqueues.PacketQueueInterface {
	return &qosConfig.config.Queues[ind]
}

func (qosConfig *QosConfiguration) GetConfig() *qosqueues.InternalRouterConfig {
	return &qosConfig.config
}

func (qosConfig *QosConfiguration) GetLegacyConfig() *qosconf.ExternalConfig {
	return &qosConfig.legacyConfig
}

// SetAndInitSchedul is necessary to set up a mock scheduler for testing. Do not use for anything else.
func (qosConfig *QosConfiguration) SetAndInitSchedul(sched qosscheduler.SchedulerInterface) {
	qosConfig.schedul = sched
	qosConfig.schedul.Init(qosConfig.config)
}

func InitQos(extConf qosconf.ExternalConfig, forwarder func(rp *rpkt.RtrPkt)) (QosConfiguration, error) {

	qConfig := QosConfiguration{}

	var err error
	if err = ConvertExternalToInternalConfig(&qConfig, extConf); err != nil {
		log.Error("Initialising the classification data structures has failed", "error", err)
	}
	if err = InitClassification(&qConfig); err != nil {
		log.Error("Initialising the classification data structures has failed", "error", err)
	}
	if err = InitScheduler(&qConfig, forwarder); err != nil {
		log.Error("Initialising the scheduler has failed", "error", err)
	}
	if err = InitWorkers(&qConfig); err != nil {
		log.Error("Initialising the workers has failed", "error", err)
	}

	return qConfig, nil
}

func ConvertExternalToInternalConfig(qConfig *QosConfiguration, extConf qosconf.ExternalConfig) error {
	var err error
	qConfig.config, err = convertExternalToInteral(extConf)
	qConfig.legacyConfig = extConf
	return err
}

func InitClassification(qConfig *QosConfiguration) error {
	qConfig.config.Rules = *qosqueues.RulesToMap(qConfig.config.Rules.RulesList)
	qConfig.config.Rules.CrCache.Init(256)

	return nil
}

func InitScheduler(qConfig *QosConfiguration, forwarder func(rp *rpkt.RtrPkt)) error {
	qConfig.notifications = make(chan *qosqueues.NPkt, maxNotificationCount)
	qConfig.Forwarder = forwarder
	// qConfig.schedul = &qosscheduler.RoundRobinScheduler{}
	qConfig.schedul = &qosscheduler.DeficitRoundRobinScheduler{}
	// qConfig.schedul = &qosscheduler.MinMaxDeficitRoundRobinScheduler{}
	// qConfig.schedul = &qosscheduler.RateRoundRobinScheduler{}
	qConfig.schedul.Init(qConfig.config)
	go qConfig.schedul.Dequeuer(qConfig.config, qConfig.Forwarder)

	return nil
}

func InitWorkers(qConfig *QosConfiguration) error {
	noWorkers := len(qConfig.config.Queues)
	qConfig.worker = workerConfiguration{noWorkers, 256}
	qConfig.workerChannels = make([]chan *qosqueues.QPkt, qConfig.worker.noWorker)

	for i := range qConfig.workerChannels {
		qConfig.workerChannels[i] = make(chan *qosqueues.QPkt, qConfig.worker.workLength)

		go worker(qConfig, &qConfig.workerChannels[i])
	}

	return nil
}

func (qosConfig *QosConfiguration) QueuePacket(rp *rpkt.RtrPkt) {

	//log.Trace("preRouteStep")
	//log.Trace("We have rules: ", "len(Rules)", len(qosConfig.GetConfig().Rules))

	rc := qosqueues.RegularClassRule{}

	// queueNo := qosqueues.GetQueueNumberForPacket(qosConfig.GetConfig(), rp)
	config := qosConfig.GetConfig()
	// queueNo := rc.GetRuleForPacket(config, rp).QueueNumber

	rule := rc.GetRuleForPacket(config, rp)

	queueNo := 0
	if rule != nil {
		queueNo = rule.QueueNumber
	}

	qp := qosqueues.QPkt{Rp: rp, QueueNo: queueNo}

	//log.Trace("Our packet is", "QPkt", qp)
	//log.Trace("Number of workers", "qosConfig.worker.noWorker", qosConfig.worker.noWorker)
	//log.Trace("Sending it to worker", "workerNo", queueNo%qosConfig.worker.noWorker)

	// log.Debug("Put packet on queue", "queueNo", queueNo)

	select {
	case *qosConfig.schedul.GetMessages() <- true:
		//log.Trace("sent message")
	default:
		//log.Trace("no message sent")
	}

	// sch := qosConfig.schedul.(*qosscheduler.DeficitRoundRobinScheduler)
	// sch.UpdateIncoming(queueNo)
	qosConfig.SendToWorker(queueNo, &qp)

	// putOnQueue(qosConfig, queueNo, &qp)

	//log.Trace("Finished QueuePacket")

}

func worker(qosConfig *QosConfiguration, workChannel *chan *qosqueues.QPkt) {
	for {
		qp := <-*workChannel
		queueNo := qp.QueueNo
		putOnQueue(qosConfig, queueNo, qp)
	}
}

func putOnQueue(qosConfig *QosConfiguration, queueNo int, qp *qosqueues.QPkt) {
	polAct := qosConfig.config.Queues[queueNo].Police(qp)
	profAct := qosConfig.config.Queues[queueNo].CheckAction()

	act := qosqueues.ReturnAction(polAct, profAct)

	switch act {
	case qosconf.PASS:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
	case qosconf.NOTIFY:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
		qosConfig.SendNotification(qp)
	case qosconf.DROPNOTIFY:
		qosConfig.dropPacket(qp.Rp)
		qosConfig.SendNotification(qp)
	case qosconf.DROP:
		qosConfig.dropPacket(qp.Rp)
	default:
		qosConfig.dropPacket(qp.Rp)
		// qosConfig.config.Queues[queueNo].Enqueue(qp)
	}
}

// SendNotification might be needed for the part of @stygerma
func (qosConfig *QosConfiguration) SendNotification(qp *qosqueues.QPkt) {
}

func (qosConfig *QosConfiguration) dropPacket(rp *rpkt.RtrPkt) {
	defer rp.Release()
	// qosConfig.notifications <- &qosqueues.NPkt{}
	qosConfig.droppedPackets++
	log.Debug("Dropping packet", "qosConfig.droppedPackets", qosConfig.droppedPackets)
	// panic("Do not drop packets")

}

func convertExternalToInteral(extConf qosconf.ExternalConfig) (qosqueues.InternalRouterConfig, error) {

	var internalRules []qosqueues.InternalClassRule
	var internalQueues []qosqueues.PacketQueueInterface

	rc := extConf

	for _, rule := range rc.ExternalRules {
		intRule, err := qosqueues.ConvClassRuleToInternal(rule)
		if err != nil {
			log.Error("Error reading config file", "error", err)
		}
		internalRules = append(internalRules, intRule)
	}

	log.Trace("Loop over Rules")
	for _, iq := range internalRules {
		log.Trace("We have gotten the rule", "rule", iq)
	}

	var intQue qosqueues.PacketQueue
	for _, extQue := range rc.ExternalQueues {

		muta := &sync.Mutex{}
		mutb := &sync.Mutex{}

		queueToUse := &qosqueues.ChannelPacketQueue{}

		intQue = convertExternalToInteralQueue(extQue)
		queueToUse.InitQueue(intQue, muta, mutb)
		internalQueues = append(internalQueues, queueToUse)
	}

	log.Trace("Loop over queues")
	for _, iq := range internalQueues {
		log.Trace("We have gotten the queue", "queue", iq.GetPacketQueue().Name)
	}

	bw := convStringToNumber(rc.SchedulerConfig.Bandwidth)

	sc := qosqueues.SchedulerConfig{Latency: rc.SchedulerConfig.Latency, Bandwidth: bw}

	return qosqueues.InternalRouterConfig{Scheduler: sc, Queues: internalQueues, Rules: qosqueues.MapRules{RulesList: internalRules}}, nil

}

func convertExternalToInteralQueue(extQueue qosconf.ExternalPacketQueue) qosqueues.PacketQueue {

	pq := qosqueues.PacketQueue{
		Name:         extQueue.Name,
		ID:           extQueue.ID,
		MinBandwidth: extQueue.MinBandwidth,
		MaxBandWidth: extQueue.MaxBandWidth,
		PoliceRate:   convStringToNumber(extQueue.PoliceRate),
		MaxLength:    extQueue.MaxLength,
		Priority:     extQueue.Priority,
		Profile:      convertActionProfiles(extQueue.Profile),
	}

	return pq
}
func convertActionProfiles(externalActionProfile []qosconf.ActionProfile) []qosqueues.ActionProfile {

	ret := make([]qosqueues.ActionProfile, 0)

	for _, prof := range externalActionProfile {
		ret = append(ret, convertActionProfile(prof))
	}
	return ret
}

func convertActionProfile(externalActionProfile qosconf.ActionProfile) qosqueues.ActionProfile {

	ap := qosqueues.ActionProfile{
		FillLevel: externalActionProfile.FillLevel,
		Prob:      externalActionProfile.Prob,
		Action:    convertPoliceAction(externalActionProfile.Action),
	}

	return ap
}

func convertPoliceAction(externalPoliceAction qosconf.PoliceAction) qosconf.PoliceAction {
	return qosconf.PoliceAction(externalPoliceAction)
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
