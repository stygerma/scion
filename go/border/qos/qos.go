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

	"github.com/scionproto/scion/go/border/qos/conf"
	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/qos/scheduler"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

const maxNotificationCount = 1024

type QosConfiguration struct {
	worker workerConfiguration

	config         queues.InternalRouterConfig
	schedul        scheduler.SchedulerInterface
	legacyConfig   conf.ExternalConfig
	notifications  chan *queues.NPkt
	workerChannels [](chan *queues.QPkt)
	Forwarder      func(rp *rpkt.RtrPkt)

	droppedPackets int
}

type workerConfiguration struct {
	noWorker   int
	workLength int
}

func (qosConfig *QosConfiguration) SendToWorker(i int, qpkt *queues.QPkt) {
	qosConfig.workerChannels[i] <- qpkt
}

func (qosConfig *QosConfiguration) GetWorkerChannels() *[](chan *queues.QPkt) {
	return &qosConfig.workerChannels
}

func (qosConfig *QosConfiguration) GetQueues() *[]queues.PacketQueueInterface {
	return &qosConfig.config.Queues
}

func (qosConfig *QosConfiguration) GetQueue(ind int) *queues.PacketQueueInterface {
	return &qosConfig.config.Queues[ind]
}

func (qosConfig *QosConfiguration) GetConfig() *queues.InternalRouterConfig {
	return &qosConfig.config
}

func (qosConfig *QosConfiguration) GetLegacyConfig() *conf.ExternalConfig {
	return &qosConfig.legacyConfig
}

// SetAndInitSchedul is necessary to set up a mock scheduler for testing. Do not use for anything else.
func (qosConfig *QosConfiguration) SetAndInitSchedul(sched scheduler.SchedulerInterface) {
	qosConfig.schedul = sched
	qosConfig.schedul.Init(qosConfig.config)
}

func InitQos(extConf conf.ExternalConfig, forwarder func(rp *rpkt.RtrPkt)) (QosConfiguration, error) {

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

func ConvertExternalToInternalConfig(qConfig *QosConfiguration, extConf conf.ExternalConfig) error {
	var err error
	qConfig.config, err = convertExternalToInteral(extConf)
	qConfig.legacyConfig = extConf
	return err
}

func InitClassification(qConfig *QosConfiguration) error {
	qConfig.config.Rules = *queues.RulesToMap(qConfig.config.Rules.RulesList)
	qConfig.config.Rules.CrCache.Init(256)

	return nil
}

func InitScheduler(qConfig *QosConfiguration, forwarder func(rp *rpkt.RtrPkt)) error {
	qConfig.notifications = make(chan *queues.NPkt, maxNotificationCount)
	qConfig.Forwarder = forwarder
	// qConfig.schedul = &scheduler.RoundRobinScheduler{}
	qConfig.schedul = &scheduler.DeficitRoundRobinScheduler{}
	// qConfig.schedul = &scheduler.MinMaxDeficitRoundRobinScheduler{}
	// qConfig.schedul = &scheduler.RateRoundRobinScheduler{}
	qConfig.schedul.Init(qConfig.config)
	go qConfig.schedul.Dequeuer(qConfig.config, qConfig.Forwarder)

	return nil
}

func InitWorkers(qConfig *QosConfiguration) error {
	noWorkers := len(qConfig.config.Queues)
	qConfig.worker = workerConfiguration{noWorkers, 256}
	qConfig.workerChannels = make([]chan *queues.QPkt, qConfig.worker.noWorker)

	for i := range qConfig.workerChannels {
		qConfig.workerChannels[i] = make(chan *queues.QPkt, qConfig.worker.workLength)

		go worker(qConfig, &qConfig.workerChannels[i])
	}

	return nil
}

func (qosConfig *QosConfiguration) QueuePacket(rp *rpkt.RtrPkt) {

	//log.Trace("preRouteStep")
	//log.Trace("We have rules: ", "len(Rules)", len(qosConfig.GetConfig().Rules))

	rc := queues.RegularClassRule{}

	// queueNo := queues.GetQueueNumberForPacket(qosConfig.GetConfig(), rp)
	config := qosConfig.GetConfig()
	// queueNo := rc.GetRuleForPacket(config, rp).QueueNumber

	rule := rc.GetRuleForPacket(config, rp)

	queueNo := 0
	if rule != nil {
		queueNo = rule.QueueNumber
	}

	qp := queues.QPkt{Rp: rp, QueueNo: queueNo}

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

	// sch := qosConfig.schedul.(*scheduler.DeficitRoundRobinScheduler)
	// sch.UpdateIncoming(queueNo)
	qosConfig.SendToWorker(queueNo, &qp)

	// putOnQueue(qosConfig, queueNo, &qp)

	//log.Trace("Finished QueuePacket")

}

func worker(qosConfig *QosConfiguration, workChannel *chan *queues.QPkt) {
	for {
		qp := <-*workChannel
		queueNo := qp.QueueNo
		putOnQueue(qosConfig, queueNo, qp)
	}
}

func putOnQueue(qosConfig *QosConfiguration, queueNo int, qp *queues.QPkt) {
	polAct := qosConfig.config.Queues[queueNo].Police(qp)
	profAct := qosConfig.config.Queues[queueNo].CheckAction()

	act := queues.ReturnAction(polAct, profAct)

	switch act {
	case conf.PASS:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
	case conf.NOTIFY:
		qosConfig.config.Queues[queueNo].Enqueue(qp)
		qosConfig.SendNotification(qp)
	case conf.DROPNOTIFY:
		qosConfig.dropPacket(qp.Rp)
		qosConfig.SendNotification(qp)
	case conf.DROP:
		qosConfig.dropPacket(qp.Rp)
	default:
		qosConfig.dropPacket(qp.Rp)
		// qosConfig.config.Queues[queueNo].Enqueue(qp)
	}
}

// SendNotification might be needed for the part of @stygerma
func (qosConfig *QosConfiguration) SendNotification(qp *queues.QPkt) {
}

func (qosConfig *QosConfiguration) dropPacket(rp *rpkt.RtrPkt) {
	defer rp.Release()
	// qosConfig.notifications <- &queues.NPkt{}
	qosConfig.droppedPackets++
	log.Debug("Dropping packet", "qosConfig.droppedPackets", qosConfig.droppedPackets)
	// panic("Do not drop packets")

}

func convertExternalToInteral(extConf conf.ExternalConfig) (queues.InternalRouterConfig, error) {

	var internalRules []queues.InternalClassRule
	var internalQueues []queues.PacketQueueInterface

	rc := extConf

	for _, rule := range rc.ExternalRules {
		intRule, err := queues.ConvClassRuleToInternal(rule)
		if err != nil {
			log.Error("Error reading config file", "error", err)
		}
		internalRules = append(internalRules, intRule)
	}

	log.Trace("Loop over Rules")
	for _, iq := range internalRules {
		log.Trace("We have gotten the rule", "rule", iq)
	}

	var intQue queues.PacketQueue
	for _, extQue := range rc.ExternalQueues {

		muta := &sync.Mutex{}
		mutb := &sync.Mutex{}

		queueToUse := &queues.ChannelPacketQueue{}

		intQue = convertExternalToInteralQueue(extQue)
		queueToUse.InitQueue(intQue, muta, mutb)
		internalQueues = append(internalQueues, queueToUse)
	}

	log.Trace("Loop over queues")
	for _, iq := range internalQueues {
		log.Trace("We have gotten the queue", "queue", iq.GetPacketQueue().Name)
	}

	bw := convStringToNumber(rc.SchedulerConfig.Bandwidth)

	sc := queues.SchedulerConfig{Latency: rc.SchedulerConfig.Latency, Bandwidth: bw}

	return queues.InternalRouterConfig{Scheduler: sc, Queues: internalQueues, Rules: queues.MapRules{RulesList: internalRules}}, nil

}

func convertExternalToInteralQueue(extQueue conf.ExternalPacketQueue) queues.PacketQueue {

	pq := queues.PacketQueue{
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
func convertActionProfiles(externalActionProfile []conf.ActionProfile) []queues.ActionProfile {

	ret := make([]queues.ActionProfile, 0)

	for _, prof := range externalActionProfile {
		ret = append(ret, convertActionProfile(prof))
	}
	return ret
}

func convertActionProfile(externalActionProfile conf.ActionProfile) queues.ActionProfile {

	ap := queues.ActionProfile{
		FillLevel: externalActionProfile.FillLevel,
		Prob:      externalActionProfile.Prob,
		Action:    convertPoliceAction(externalActionProfile.Action),
	}

	return ap
}

func convertPoliceAction(externalPoliceAction conf.PoliceAction) conf.PoliceAction {
	return conf.PoliceAction(externalPoliceAction)
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
