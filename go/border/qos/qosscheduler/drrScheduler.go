package qosscheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

type DeficitRoundRobinScheduler struct {
	quantumSum       int
	totalLength      int
	messages         chan bool
	totalQueueLength int

	timeToSleep int
	tb          qosqueues.TokenBucket
}

var _ SchedulerInterface = (*DeficitRoundRobinScheduler)(nil)

func (sched *DeficitRoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	for i := 0; i < len(routerConfig.Queues); i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}

	if len(routerConfig.Queues) == 5 {
		log.Debug("Priorities", "0", routerConfig.Queues[0].GetPriority(), "1", routerConfig.Queues[1].GetPriority(), "2", routerConfig.Queues[2].GetPriority())

		sched.tb.Init(1250000) // 10 Mbit
	} else {
		sched.tb.Init(125000000) // 1000 Mbit
	}

}

func getNoPacketsToDequeue(totalLength int, priority int, totalPriority int) int {
	// return int(math.Floor(float64(totalLength) / float64(totalPriority) * float64(priority)))
	// return totalLength / totalPriority * priority
	return priority
}

var incoming [5]int
var lastRound [5]int
var attempted [5]int
var total [5]int

func (sched *DeficitRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	nopkts := getNoPacketsToDequeue(sched.totalQueueLength, queue.GetPriority(), sched.quantumSum)
	pktToDequeue := nopkts

	var qp *qosqueues.QPkt

	attempted[queueNo] += pktToDequeue

	for i := 0; i < pktToDequeue; i++ {

		qp = queue.Pop()

		if qp == nil {
			break
		}

		for !(sched.tb.Take(qp.Rp.Bytes().Len())) {
			time.Sleep(50 * time.Millisecond)
		}

		lastRound[queueNo]++
		total[queueNo]++
		forwarder(qp.Rp)
	}
}

func (sched *DeficitRoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		sched.totalQueueLength = 0
		for i := 0; i < sched.totalLength; i++ {
			sched.totalQueueLength += routerConfig.Queues[i].GetLength()
		}

		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}

		sched.showLog(routerConfig)

	}
}

func (sched *DeficitRoundRobinScheduler) UpdateIncoming(queueNo int) {
	incoming[queueNo]++
}

var iterations int

func (sched *DeficitRoundRobinScheduler) showLog(routerConfig qosqueues.InternalRouterConfig) {

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

func (sched *DeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
