package qosscheduler

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is also a deficit round robin dequeuer. But instead of the priority field it uses the min-bandwidth field for the minimum number of packets to dequeue. If there are fewer than the minimal value of packets to dequeue, the remaining min-bandwidth will be put onto a surplus counter and another queue might use more than its min-bandwidth (but still less than its max-bandwidth).

type MinMaxDeficitRoundRobinScheduler struct {
	quantumSum          int
	totalLength         int
	schedulerSurplus    surplus
	schedulerSurplusMtx *sync.Mutex
	messages            chan bool
	jobs                chan int

	tb qosqueues.TokenBucket
}

type surplus struct {
	Surplus    int
	Payments   []int
	MaxSurplus int
}

var _ SchedulerInterface = (*MinMaxDeficitRoundRobinScheduler)(nil)

// var jobs chan int

func (sched *MinMaxDeficitRoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	sched.schedulerSurplusMtx = &sync.Mutex{}
	sched.schedulerSurplus = surplus{0, make([]int, sched.totalLength), -1}

	sched.jobs = make(chan int, sched.totalLength)

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetMinBandwidth()
	}
	for i := 0; i < sched.totalLength; i++ {
		sched.schedulerSurplus.MaxSurplus += routerConfig.Queues[i].GetMinBandwidth()
	}

	if len(routerConfig.Queues) == 5 {
		log.Debug("Priorities", "0", routerConfig.Queues[0].GetPriority(), "1", routerConfig.Queues[1].GetPriority(), "2", routerConfig.Queues[2].GetPriority())

		sched.tb.Init(2 * 1250000) // 20 Mbit
	} else {
		sched.tb.Init(125000000) // 1000 Mbit
	}

}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		for i := 0; i < sched.totalLength; i++ {
			_ = <-sched.jobs
		}
		sched.LogUpdate(routerConfig)
	}
}

// var attemptedToRoute [5]int
// var packets [5]int
// var surplusAdjustments [5]int
var t0 time.Time

func (sched *MinMaxDeficitRoundRobinScheduler) LogUpdate(routerConfig qosqueues.InternalRouterConfig) {

	iterations++
	if time.Now().Sub(t0) > time.Duration(5*time.Second) {

		var queLen [5]int
		for i := 0; i < sched.totalLength; i++ {
			queLen[i] = routerConfig.Queues[i].GetLength()
		}
		log.Debug("STAT", "iterations", iterations, "incoming", incoming, "deqLastRound", lastRound, "deqAttempted", attempted, "deqTotal", total, "currQueueLen", queLen)
		for i := 0; i < len(lastRound); i++ {
			lastRound[i] = 0
		}
		for i := 0; i < len(attempted); i++ {

			attempted[i] = 0
		}
		for i := 0; i < len(incoming); i++ {
			incoming[i] = 0
		}
		t0 = time.Now()
		iterations = 0
	}

}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	pktToDequeue := sched.adjustForQuantum(queue)
	pktToDequeue = sched.adjustForSurplus(queue, pktToDequeue, queueNo)

	attempted[queueNo] += pktToDequeue

	sched.dequeuePackets(queue, pktToDequeue, forwarder, queueNo)

}

func (sched *MinMaxDeficitRoundRobinScheduler) dequeuePackets(queue qosqueues.PacketQueueInterface, pktToDequeue int, forwarder func(rp *rpkt.RtrPkt), queueNo int) int {
	var qp *qosqueues.QPkt
	j := 0
	for i := 0; i < pktToDequeue; i++ {
		qp = queue.Pop()
		if qp == nil {
			break
		}

		for !(sched.tb.Take(qp.Rp.Bytes().Len())) {
			time.Sleep(50 * time.Millisecond)
		}

		j++
		forwarder(qp.Rp)
	}
	lastRound[queueNo] += j
	total[queueNo] += j
	sched.jobs <- j
	return j
}

func (sched *MinMaxDeficitRoundRobinScheduler) adjustForSurplus(queue qosqueues.PacketQueueInterface, pktToDequeue int, queueNo int) int {

	length := queue.GetLength()

	if sched.surplusAvailable() {
		if length > pktToDequeue {
			pktToDequeue = sched.getFromSurplus(queue, queueNo, length)
		}
	} else {
		if pktToDequeue-length > 0 {
			sched.payIntoSurplus(queue, queueNo, min(pktToDequeue-length, queue.GetMinBandwidth()))
		}
	}

	return pktToDequeue
}

func (sched *MinMaxDeficitRoundRobinScheduler) adjustForQuantum(queue qosqueues.PacketQueueInterface) int {

	a := queue.GetMinBandwidth()

	b := 10.0 / float64(sched.quantumSum)
	pd := float64(a) * b
	pktToDequeue := max(int(pd), 1)

	return pktToDequeue
}

func (sched *MinMaxDeficitRoundRobinScheduler) getFromSurplus(queue qosqueues.PacketQueueInterface, queueNo int, request int) int {

	maxAllowedTakeout := queue.GetMaxBandwidth()
	maxRequestedTakeout := request - queue.GetMinBandwidth()
	maxTakeout := min(maxRequestedTakeout, maxAllowedTakeout)

	credit := min(sched.schedulerSurplus.Surplus, maxTakeout)
	credit = min(credit+queue.GetMinBandwidth(), queue.GetMaxBandwidth())
	sched.schedulerSurplus.Surplus -= (credit - queue.GetMinBandwidth())
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus

	return credit

}

func (sched *MinMaxDeficitRoundRobinScheduler) payIntoSurplus(queue qosqueues.PacketQueueInterface, queueNo int, payment int) {

	sched.schedulerSurplus.Surplus = min(sched.schedulerSurplus.Surplus+payment, sched.schedulerSurplus.MaxSurplus)
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus
}

func (sched *MinMaxDeficitRoundRobinScheduler) surplusAvailable() bool {

	return sched.schedulerSurplus.Surplus > 0
}

func (sched *MinMaxDeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
