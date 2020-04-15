package scheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

type DeficitRoundRobinScheduler struct {
	quantumSum       int
	totalLength      int
	messages         chan bool
	totalQueueLength int
	sleepDuration    int
	tb               queues.TokenBucket
}

var _ SchedulerInterface = (*DeficitRoundRobinScheduler)(nil)

func (sched *DeficitRoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	for i := 0; i < len(routerConfig.Queues); i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}
	sched.tb.Init(routerConfig.Scheduler.Bandwidth)
	sched.sleepDuration = routerConfig.Scheduler.Latency
}

func getNoPacketsToDequeue(totalLength int, priority int, totalPriority int) int {
	return priority
}

var incoming [5]int
var lastRound [5]int
var attempted [5]int
var total [5]int

func (sched *DeficitRoundRobinScheduler) Dequeue(queue queues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	nopkts := getNoPacketsToDequeue(sched.totalQueueLength, queue.GetPriority(), sched.quantumSum)
	pktToDequeue := nopkts

	var qp *queues.QPkt

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

func (sched *DeficitRoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		t0 := time.Now()
		sched.totalQueueLength = 0
		for i := 0; i < sched.totalLength; i++ {
			sched.totalQueueLength += routerConfig.Queues[i].GetLength()
		}

		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}

		sched.showLog(routerConfig)

		for time.Now().Sub(t0) < time.Duration(time.Duration(sched.sleepDuration)*time.Microsecond) {
			time.Sleep(time.Duration(sched.sleepDuration/10) * time.Microsecond)
		}
	}
}

func (sched *DeficitRoundRobinScheduler) UpdateIncoming(queueNo int) {
	incoming[queueNo]++
}

var iterations int

func (sched *DeficitRoundRobinScheduler) showLog(routerConfig queues.InternalRouterConfig) {

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
