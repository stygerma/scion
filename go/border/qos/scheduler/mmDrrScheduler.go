package scheduler

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is also a deficit round robin dequeuer. But instead of the priority field it
// uses the min-bandwidth field for the minimum number of packets to dequeue. If there are
// fewer than the minimal value of packets to dequeue, the remaining min-bandwidth will be
// put onto a surplus counter and another queue might use more than its min-bandwidth
// (but still less than its max-bandwidth).

type MinMaxDeficitRoundRobinScheduler struct {
	quantumSum          int
	totalLength         int
	schedulerSurplus    surplus
	schedulerSurplusMtx *sync.Mutex
	messages            chan bool
	jobs                chan int

	sleepDuration int
	tb            queues.TokenBucket

	logger ScheduleLogger
}

type surplus struct {
	Surplus    int
	Payments   []int
	MaxSurplus int
}

var _ SchedulerInterface = (*MinMaxDeficitRoundRobinScheduler)(nil)

func (sched *MinMaxDeficitRoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	sched.logger = initLogger(sched.totalLength)

	sched.schedulerSurplusMtx = &sync.Mutex{}
	sched.schedulerSurplus = surplus{0, make([]int, sched.totalLength), -1}

	sched.jobs = make(chan int, sched.totalLength)

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetMinBandwidth()
	}
	for i := 0; i < sched.totalLength; i++ {
		sched.schedulerSurplus.MaxSurplus += routerConfig.Queues[i].GetMinBandwidth()
	}

	sched.tb.Init(routerConfig.Scheduler.Bandwidth)
	sched.sleepDuration = routerConfig.Scheduler.Latency
}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig,
	forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		sleepDuration := time.Duration(time.Duration(sched.sleepDuration) * time.Microsecond)
		t0 := time.Now()
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		for i := 0; i < sched.totalLength; i++ {
			_ = <-sched.jobs
		}
		sched.LogUpdate(routerConfig)

		for time.Now().Sub(t0) < sleepDuration {
			time.Sleep(time.Duration(sched.sleepDuration/10) * time.Microsecond)
		}
	}
}

func (sched *MinMaxDeficitRoundRobinScheduler) LogUpdate(
	routerConfig queues.InternalRouterConfig) {

	sched.logger.iterations++
	if time.Now().Sub(sched.logger.t0) > time.Duration(5*time.Second) {

		var queLen [5]int
		for i := 0; i < sched.totalLength; i++ {
			queLen[i] = routerConfig.Queues[i].GetLength()
		}
		log.Debug("STAT",
			"iterations", sched.logger.iterations,
			"incoming", sched.logger.incoming,
			"deqLastRound", sched.logger.lastRound,
			"deqAttempted", sched.logger.attempted,
			"deqTotal", sched.logger.total, "currQueueLen", queLen)
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

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeue(queue queues.PacketQueueInterface,
	forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	pktToDequeue := sched.adjustForQuantum(queue)
	pktToDequeue = sched.adjustForSurplus(queue, pktToDequeue, queueNo)

	sched.logger.attempted[queueNo] += pktToDequeue

	sched.dequeuePackets(queue, pktToDequeue, forwarder, queueNo)

}

func (sched *MinMaxDeficitRoundRobinScheduler) dequeuePackets(queue queues.PacketQueueInterface,
	pktToDequeue int, forwarder func(rp *rpkt.RtrPkt), queueNo int) int {
	var qp *queues.QPkt
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
	sched.logger.lastRound[queueNo] += j
	sched.logger.total[queueNo] += j
	sched.jobs <- j
	return j
}

func (sched *MinMaxDeficitRoundRobinScheduler) adjustForSurplus(queue queues.PacketQueueInterface,
	pktToDequeue int, queueNo int) int {

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

func (sched *MinMaxDeficitRoundRobinScheduler) adjustForQuantum(
	queue queues.PacketQueueInterface) int {

	a := queue.GetMinBandwidth()

	b := 10.0 / float64(sched.quantumSum)
	pd := float64(a) * b
	pktToDequeue := max(int(pd), 1)

	return pktToDequeue
}

func (sched *MinMaxDeficitRoundRobinScheduler) getFromSurplus(queue queues.PacketQueueInterface,
	queueNo int, request int) int {

	maxAllowedTakeout := queue.GetMaxBandwidth()
	maxRequestedTakeout := request - queue.GetMinBandwidth()
	maxTakeout := min(maxRequestedTakeout, maxAllowedTakeout)

	credit := min(sched.schedulerSurplus.Surplus, maxTakeout)
	credit = min(credit+queue.GetMinBandwidth(), queue.GetMaxBandwidth())
	sched.schedulerSurplus.Surplus -= (credit - queue.GetMinBandwidth())
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus

	return credit

}

func (sched *MinMaxDeficitRoundRobinScheduler) payIntoSurplus(
	queue queues.PacketQueueInterface, queueNo int, payment int) {
	a := sched.schedulerSurplus.Surplus + payment
	b := sched.schedulerSurplus.MaxSurplus
	sched.schedulerSurplus.Surplus = min(a, b)
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus
}

func (sched *MinMaxDeficitRoundRobinScheduler) surplusAvailable() bool {

	return sched.schedulerSurplus.Surplus > 0
}

func (sched *MinMaxDeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
