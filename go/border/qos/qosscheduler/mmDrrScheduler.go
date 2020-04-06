package qosscheduler

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// This is also a deficit round robin dequeuer. But instead of the priority field it uses the min-bandwidth field for the minimum number of packets to dequeue. If there are fewer than the minimal value of packets to dequeue, the remaining min-bandwidth will be put onto a surplus counter and another queue might use more than its min-bandwidth (but still less than its max-bandwidth).

type MinMaxDeficitRoundRobinScheduler struct {
	quantumSum          int
	totalLength         int
	schedulerSurplus    surplus
	schedulerSurplusMtx *sync.Mutex
	messages            chan bool
	jobs                chan int
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

}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	// t0 := time.Now()
	for {
		// t0 = time.Now()
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		for i := 0; i < sched.totalLength; i++ {
			_ = <-sched.jobs
			// log.Debug("Dequed", "numberOfPackets", j)
		}
		time.Sleep(500 * time.Nanosecond)
		// t1 := time.Now()
		// log.Debug("One round dequeued in", "t1.Sub(t0)", t1.Sub(t0))
	}
}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	// length := queue.GetLength()
	// log.Debug("The queue has length", "queueNo", queueNo, "length", length)

	pktToDequeue := sched.adjustForQuantum(queue)

	// log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	pktToDequeue = sched.adjustForSurplus(queue, pktToDequeue, queueNo)

	// log.Debug("Dequeue packets from queue", "queueNo", queueNo)
	go sched.dequeuePackets(queue, pktToDequeue, forwarder)
	// sched.dequeuePackets(queue, pktToDequeue, forwarder)

	// log.Debug("Dequeued packets from queue", "queueNo", queueNo, "toDeqLength", pktToDequeue, "totalLength", length)

}

func (sched *MinMaxDeficitRoundRobinScheduler) dequeuePackets(queue qosqueues.PacketQueueInterface, pktToDequeue int, forwarder func(rp *rpkt.RtrPkt)) int {
	var qp *qosqueues.QPkt
	j := 0
	for i := 0; i < pktToDequeue; i++ {
		qp = queue.Pop()
		if qp == nil {
			break
		}
		j++
		forwarder(qp.Rp)
	}
	// fmt.Println("before sending to jobs")
	sched.jobs <- j
	// fmt.Println("after sending to jobs")
	return j
}

func (sched *MinMaxDeficitRoundRobinScheduler) adjustForSurplus(queue qosqueues.PacketQueueInterface, pktToDequeue int, queueNo int) int {

	length := queue.GetLength()

	if sched.surplusAvailable() {
		// log.Debug("Surplus available", "surplus", sched.schedulerSurplus)
		if length > pktToDequeue {
			pktToDequeue = sched.getFromSurplus(queue, queueNo, length)
			// log.Debug("Dequeueing above minimum", "quantum", pktToDequeue)
		}
	} else {
		if pktToDequeue-length > 0 {
			sched.payIntoSurplus(queue, queueNo, min(pktToDequeue-length, queue.GetMinBandwidth()))
			// log.Debug("Paying into surplus", "payment", pktToDequeue-length)
		}
	}

	// log.Debug("Queue got surplus", "queueNo", queueNo, "pktToDequeue", pktToDequeue)

	return pktToDequeue
}

func (sched *MinMaxDeficitRoundRobinScheduler) adjustForQuantum(queue qosqueues.PacketQueueInterface) int {

	a := queue.GetMinBandwidth()
	// fmt.Println("a is", a)
	// TODO: We need the total queue length instead of 100 here
	b := 100.0 / float64(sched.quantumSum)
	// b := float64(sched.currLength) / float64(sched.quantumSum)
	// fmt.Println("b is", b)
	// pd := float64(queue.GetPriority()) * b
	pd := float64(a) * b
	// fmt.Println("pd is", pd)
	pktToDequeue := max(int(pd), 1)
	// fmt.Println("pktToDequeue is", pktToDequeue)

	return pktToDequeue

}

func (sched *MinMaxDeficitRoundRobinScheduler) getFromSurplus(queue qosqueues.PacketQueueInterface, queueNo int, request int) int {

	// sched.schedulerSurplusMtx.Lock()
	// defer sched.schedulerSurplusMtx.Unlock()

	// Check limit for queue
	// Take out of surplus

	// maxAllowedTakeout := queue.GetMaxBandwidth() - queue.GetMinBandwidth()
	maxAllowedTakeout := queue.GetMaxBandwidth()
	maxRequestedTakeout := request - queue.GetMinBandwidth()
	// maxRequestedTakeout := request
	maxTakeout := min(maxRequestedTakeout, maxAllowedTakeout)

	// fmt.Println("maxAllowedTakeout", maxAllowedTakeout, "maxRequestedTakeout", maxRequestedTakeout, "maxTakeout", maxTakeout)

	credit := min(sched.schedulerSurplus.Surplus, maxTakeout)
	credit = min(credit+queue.GetMinBandwidth(), queue.GetMaxBandwidth())
	sched.schedulerSurplus.Surplus -= (credit - queue.GetMinBandwidth())
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus

	// log.Debug("Queue got credit", "queueNo", queueNo, "credit", credit)

	return credit

}

func (sched *MinMaxDeficitRoundRobinScheduler) payIntoSurplus(queue qosqueues.PacketQueueInterface, queueNo int, payment int) {

	// sched.schedulerSurplusMtx.Lock()
	// defer sched.schedulerSurplusMtx.Unlock()

	// fmt.Println("Max", sched.schedulerSurplus.MaxSurplus)
	// fmt.Println("Set to", min(sched.schedulerSurplus.Surplus+payment, sched.schedulerSurplus.MaxSurplus))
	sched.schedulerSurplus.Surplus = min(sched.schedulerSurplus.Surplus+payment, sched.schedulerSurplus.MaxSurplus)
	// fmt.Println("Actual", sched.schedulerSurplus.Surplus)
	sched.schedulerSurplus.Payments[queueNo] = sched.schedulerSurplus.Surplus
}

func (sched *MinMaxDeficitRoundRobinScheduler) surplusAvailable() bool {

	// sched.schedulerSurplusMtx.Lock()
	// defer sched.schedulerSurplusMtx.Unlock()

	return sched.schedulerSurplus.Surplus > 0
}

func (sched *MinMaxDeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
