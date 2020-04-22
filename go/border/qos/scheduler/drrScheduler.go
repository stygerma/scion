package scheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// This is a deficit round robin dequeuer.
// Queues with higher priority will have more packets dequeued at the same time.

type DeficitRoundRobinScheduler struct {
	quantumSum       int
	totalLength      int
	messages         chan bool
	totalQueueLength int
	sleepDuration    int
	tb               queues.TokenBucket
	logger           ScheduleLogger
}

var _ SchedulerInterface = (*DeficitRoundRobinScheduler)(nil)

func (sched *DeficitRoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	sched.logger = initLogger(sched.totalLength)

	for i := 0; i < len(routerConfig.Queues); i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}
	sched.tb.Init(routerConfig.Scheduler.Bandwidth)
	sched.sleepDuration = routerConfig.Scheduler.Latency
}

func getNoPacketsToDequeue(totalLength int, priority int, totalPriority int) int {
	return priority
}

func (sched *DeficitRoundRobinScheduler) Dequeue(queue queues.PacketQueueInterface,
	forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	nopkts := getNoPacketsToDequeue(sched.totalQueueLength, queue.GetPriority(), sched.quantumSum)
	pktToDequeue := nopkts

	var qp *queues.QPkt

	sched.logger.attempted[queueNo] += pktToDequeue

	for i := 0; i < pktToDequeue; i++ {

		qp = queue.Pop()

		if qp == nil {
			break
		}

		for !(sched.tb.Take(qp.Rp.Bytes().Len())) {
			time.Sleep(50 * time.Millisecond)
		}

		sched.logger.lastRound[queueNo]++
		sched.logger.total[queueNo]++
		forwarder(qp.Rp)
	}
}

func (sched *DeficitRoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig,
	forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	sleepDuration := time.Duration(time.Duration(sched.sleepDuration) * time.Microsecond)
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

		for time.Now().Sub(t0) < sleepDuration {
			time.Sleep(time.Duration(sched.sleepDuration/10) * time.Microsecond)
		}
	}
}

func (sched *DeficitRoundRobinScheduler) UpdateIncoming(queueNo int) {
	sched.logger.incoming[queueNo]++
}

func (sched *DeficitRoundRobinScheduler) showLog(routerConfig queues.InternalRouterConfig) {

	sched.logger.iterations++
	if time.Now().Sub(sched.logger.t0) > time.Duration(5*time.Second) {

		var queLen [5]int
		for i := 0; i < sched.totalLength; i++ {
			queLen[i] = routerConfig.Queues[i].GetLength()
		}
		// log.Debug("STAT",
		// 	"iterations", sched.logger.iterations,
		// 	"incoming", sched.logger.incoming,
		// 	"deqLastRound",
		// 	sched.logger.lastRound, "deqAttempted",
		// 	sched.logger.attempted, "deqTotal",
		// 	sched.logger.total, "currQueueLen", queLen)
		for i := 0; i < len(sched.logger.lastRound); i++ {
			sched.logger.lastRound[i] = 0
		}
		for i := 0; i < len(sched.logger.attempted); i++ {

			sched.logger.attempted[i] = 0
		}
		for i := 0; i < len(sched.logger.incoming); i++ {
			sched.logger.incoming[i] = 0
		}
		sched.logger.t0 = time.Now()
		sched.logger.iterations = 0
	}

}

func (sched *DeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
