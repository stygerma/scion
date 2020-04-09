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
			// log.Debug("Dequed", "numberOfPackets", j, "queueNo", i)
		}
		LogUpdate(routerConfig)

		if len(routerConfig.Queues) == 3 {
			time.Sleep(400 * time.Microsecond)
		} else {
			time.Sleep(500 * time.Nanosecond)
		}

		// time.Sleep(1 * time.Millisecond)
		// t1 := time.Now()
		// log.Debug("One round dequeued in", "t1.Sub(t0)", t1.Sub(t0))
	}
}

var attemptedToRoute [3]int
var packets [3]int
var surplusAdjustments [3]int
var t0 time.Time

func LogUpdate(routerConfig qosqueues.InternalRouterConfig) {
	if time.Now().Sub(t0) > time.Duration(1*time.Second) {
		if len(routerConfig.Queues) == 3 {
			log.Debug("Queuelengths;NoDequeued", "0", routerConfig.Queues[0].GetLength(), "1", routerConfig.Queues[1].GetLength(), "2", routerConfig.Queues[2].GetLength(), "0", packets[0], "1", packets[1], "2", packets[2], "0", attemptedToRoute[0], "1", attemptedToRoute[1], "2", attemptedToRoute[2])
		}
		// log.Debug("Attempted packets w/o surplus", "attemptedToRoute", attemptedToRoute)
		// log.Debug("Ratio queue 1 to queue 2", "", float64(attemptedToRoute[1])/float64(attemptedToRoute[2]))
		// log.Debug("Deqeued packets w/o surplus", "packets", packets)
		// log.Debug("Ratio queue 1 to queue 2", "", float64(packets[1])/float64(packets[2]))
		// log.Debug("Surplus adjustments", "surplusAdjustments", surplusAdjustments)
		// log.Debug("Ratio queue 1 to queue 2", "", float64(packets[1])/float64(packets[2]))
		packets[0] = 0
		packets[1] = 0
		packets[2] = 0

		surplusAdjustments[0] = 0
		surplusAdjustments[1] = 0
		surplusAdjustments[2] = 0
		t0 = time.Now()
	}
}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	// length := queue.GetLength()
	// log.Debug("The queue has length", "queueNo", queueNo, "length", length)

	pktToDequeue := sched.adjustForQuantum(queue)

	attemptedToRoute[queueNo] += pktToDequeue

	// log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	// spsAdjust := sched.adjustForSurplus(queue, pktToDequeue, queueNo)

	// surplusAdjustments[queueNo] += spsAdjust

	// pktToDequeue = spsAdjust

	// log.Debug("Dequeue packets from queue", "queueNo", queueNo)
	go sched.dequeuePackets(queue, pktToDequeue, forwarder, queueNo)
	// sched.dequeuePackets(queue, pktToDequeue, forwarder)

	// log.Debug("Dequeued packets from queue", "queueNo", queueNo, "toDeqLength", pktToDequeue, "totalLength", length)

}

func (sched *MinMaxDeficitRoundRobinScheduler) dequeuePackets(queue qosqueues.PacketQueueInterface, pktToDequeue int, forwarder func(rp *rpkt.RtrPkt), queueNo int) int {
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
	packets[queueNo] += j
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
