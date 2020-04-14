package qosscheduler

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is also a deficit round robin dequeuer. But instead of the priority field it uses the min-bandwidth field for the minimum number of packets to dequeue. If there are fewer than the minimal value of packets to dequeue, the remaining min-bandwidth will be put onto a surplus counter and another queue might use more than its min-bandwidth (but still less than its max-bandwidth).

type RateRoundRobinScheduler struct {
	quantumSum          int
	totalLength         int
	schedulerSurplus    surplus
	schedulerSurplusMtx *sync.Mutex
	messages            chan bool
	jobs                chan int

	cirBuckets []qosqueues.TokenBucket
	pirBuckets []qosqueues.TokenBucket

	tb qosqueues.TokenBucket
}

var _ SchedulerInterface = (*RateRoundRobinScheduler)(nil)

func (sched *RateRoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	sched.schedulerSurplus = surplus{0, make([]int, sched.totalLength), -1}
	sched.schedulerSurplusMtx = &sync.Mutex{}

	sched.jobs = make(chan int, sched.totalLength)

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetMinBandwidth()
	}

	maxBW := 2000000 // 2MB

	if len(routerConfig.Queues) == 5 {
		log.Debug("Priorities", "0", routerConfig.Queues[0].GetPriority(), "1", routerConfig.Queues[1].GetPriority(), "2", routerConfig.Queues[2].GetPriority())

		sched.tb.Init(maxBW) // 20 Mbit
	} else {
		sched.tb.Init(125000000) // 1000 Mbit
	}

	sched.schedulerSurplus.MaxSurplus = maxBW

	sched.cirBuckets = make([]qosqueues.TokenBucket, sched.totalLength)
	sched.pirBuckets = make([]qosqueues.TokenBucket, sched.totalLength)

	for i := 0; i < sched.totalLength; i++ {
		// bw := float64(routerConfig.Queues[i].GetMinBandwidth()) / float64(sched.quantumSum)
		bw := float64(routerConfig.Queues[i].GetMinBandwidth()) / 100.0
		log.Debug("Init bucket with", "int(maxBW * bw)", int(float64(maxBW)*bw), "bw", bw)
		sched.cirBuckets[i].Init(int(float64(maxBW) * bw))
		// sched.cirBuckets[i].Init(maxBW)
	}
	for i := 0; i < sched.totalLength; i++ {
		// bw := float64(routerConfig.Queues[i].GetMaxBandwidth()) / float64(sched.quantumSum)
		bw := float64(routerConfig.Queues[i].GetMaxBandwidth()) / 100.0
		log.Debug("Init bucket with", "int(maxBW * bw)", int(float64(maxBW)*bw), "bw", bw)
		sched.pirBuckets[i].Init(int(float64(maxBW) * bw))
	}

}

func (sched *RateRoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		t0 := time.Now()
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		for i := 0; i < sched.totalLength; i++ {
			_ = <-sched.jobs
		}
		sched.LogUpdate(routerConfig)

		for time.Now().Sub(t0) < time.Duration(250*time.Microsecond) {
			time.Sleep(100 * time.Microsecond)
		}
	}
}

var tokensUsed [5]int
var cirTokens [5]int
var pirTokens [5]int
var payedIntoSurplus [5]int
var forceTake [5]int
var overallTokensUsed int

func (sched *RateRoundRobinScheduler) LogUpdate(routerConfig qosqueues.InternalRouterConfig) {

	iterations++
	if time.Now().Sub(t0) > time.Duration(5*time.Second) {

		var queLen [5]int
		for i := 0; i < sched.totalLength; i++ {
			queLen[i] = routerConfig.Queues[i].GetLength()
		}
		log.Debug("STAT", "iterations", iterations, "incoming", incoming, "deqLastRound", lastRound, "deqAttempted", attempted, "deqTotal", total, "currQueueLen", queLen, "surplus", sched.schedulerSurplus.Surplus)
		if len(sched.cirBuckets) > 3 {
			log.Debug("STAT Available cirTokens", "1", sched.cirBuckets[1].GetAvailable(), "2", sched.cirBuckets[2].GetAvailable(), "3", sched.cirBuckets[3].GetAvailable())
			log.Debug("STAT Available pirTokens", "1", sched.pirBuckets[1].GetAvailable(), "2", sched.pirBuckets[2].GetAvailable(), "3", sched.pirBuckets[3].GetAvailable())
			log.Debug("STAT", "tokensUsed", tokensUsed, "forceTake", forceTake, "cirTokens", cirTokens, "pirTokens", pirTokens, "payedIntoSurplus", payedIntoSurplus)
			queue0 := float64(tokensUsed[0]+forceTake[0]) / 5.0 / 1000000.0
			queue1 := float64(tokensUsed[1]+forceTake[1]) / 5.0 / 1000000.0
			queue2 := float64(tokensUsed[2]+forceTake[2]) / 5.0 / 1000000.0
			queue3 := float64(tokensUsed[3]+forceTake[3]) / 5.0 / 1000000.0
			queue4 := float64(tokensUsed[4]+forceTake[4]) / 5.0 / 1000000.0
			overall := float64(overallTokensUsed) / 5.0 / 1000000.0
			log.Debug("STAT", "overall", overall, "maxOverall", 2, "0", queue0, "1", queue1, "2", queue2, "3", queue3, "4", queue4)

		}
		for i := 0; i < len(lastRound); i++ {
			lastRound[i] = 0
		}
		for i := 0; i < len(attempted); i++ {

			attempted[i] = 0
		}
		for i := 0; i < len(incoming); i++ {
			incoming[i] = 0
		}
		for i := 0; i < len(tokensUsed); i++ {
			tokensUsed[i] = 0
		}
		for i := 0; i < len(cirTokens); i++ {
			cirTokens[i] = 0
		}
		for i := 0; i < len(pirTokens); i++ {
			pirTokens[i] = 0
		}
		for i := 0; i < len(payedIntoSurplus); i++ {
			payedIntoSurplus[i] = 0
		}
		for i := 0; i < len(forceTake); i++ {
			forceTake[i] = 0
		}
		overallTokensUsed = 0
		t0 = time.Now()
		iterations = 0
	}

}

func (sched *RateRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {
	no := queue.GetMinBandwidth()
	attempted[queueNo] += no
	sched.dequeuePackets(queue, no, forwarder, queueNo)
}

func (sched *RateRoundRobinScheduler) dequeuePackets(queue qosqueues.PacketQueueInterface, pktToDequeue int, forwarder func(rp *rpkt.RtrPkt), queueNo int) int {
	var qp *qosqueues.QPkt
	j := 0

	for i := 0; i < pktToDequeue; i++ {

		if !sched.availableFromBuckets(1500, queueNo) {
			break
		}

		qp = queue.Pop()
		if qp == nil {
			break
		}
		// log.Debug("Packet size is", "qp.Rp.Bytes().Len()", qp.Rp.Bytes().Len(), "queueNo", queueNo)
		j++
		if !(sched.takeFromBuckets(qp.Rp.Bytes().Len(), queueNo)) {
			// log.Debug("We had to force take!", "queueNO", queueNo)
			sched.cirBuckets[queueNo].ForceTake(qp.Rp.Bytes().Len())
			forwarder(qp.Rp)
			break
		}
		forwarder(qp.Rp)
	}

	lastRound[queueNo] += j
	total[queueNo] += j
	sched.jobs <- j
	for i := 0; i < 5; i++ {
		if sched.cirBuckets[queueNo].Take(1500) {
			payedIntoSurplus[queueNo] += 1500
			sched.payIntoSurplus(queue, queueNo, 1500)
		}
	}
	return j
}

// TODO: This does not respect the overall information rate
func (sched *RateRoundRobinScheduler) availableFromBuckets(packetLength int, queueNo int) bool {

	if sched.tb.Available(packetLength) {
		if sched.cirBuckets[queueNo].Available(packetLength) {
			return true
		}

		if sched.pirBuckets[queueNo].Available(packetLength) {
			if sched.availableSurplus(packetLength) {
				return true
			}
		}
	}

	return false
}

func (sched *RateRoundRobinScheduler) takeFromBuckets(packetLength int, queueNo int) bool {

	if sched.tb.Available(packetLength) {

		if sched.cirBuckets[queueNo].Take(packetLength) {
			sched.pirBuckets[queueNo].ForceTake(packetLength)
			sched.tb.Take(packetLength)
			overallTokensUsed += packetLength
			tokensUsed[queueNo] += packetLength
			cirTokens[queueNo] += packetLength
			return true
		}

		if sched.pirBuckets[queueNo].Available(packetLength) {
			if sched.takeSurplus(packetLength) {
				sched.pirBuckets[queueNo].Take(packetLength)
				sched.tb.Take(packetLength)
				overallTokensUsed += packetLength
				tokensUsed[queueNo] += packetLength
				pirTokens[queueNo] += packetLength
				return true
			}
		}
	}

	forceTake[queueNo] += packetLength

	return false
}
func (sched *RateRoundRobinScheduler) availableSurplus(amount int) bool {
	// sched.schedulerSurplusMtx.Lock()
	// defer sched.schedulerSurplusMtx.Unlock()

	if sched.schedulerSurplus.Surplus > amount {
		return true
	}
	return false
}

func (sched *RateRoundRobinScheduler) takeSurplus(amount int) bool {
	// sched.schedulerSurplusMtx.Lock()
	// defer sched.schedulerSurplusMtx.Unlock()

	if sched.schedulerSurplus.Surplus > amount {
		sched.schedulerSurplus.Surplus -= amount
		return true
	}
	return false
}

func (sched *RateRoundRobinScheduler) payIntoSurplus(queue qosqueues.PacketQueueInterface, queueNo int, payment int) {

	// sched.schedulerSurplusMtx.Lock()
	// defer sched.schedulerSurplusMtx.Unlock()

	sched.schedulerSurplus.Surplus = min(sched.schedulerSurplus.Surplus+payment, sched.schedulerSurplus.MaxSurplus)
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus
}

func (sched *RateRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
