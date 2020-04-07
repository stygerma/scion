package scheduler

import (
	"sync"

	"github.com/scionproto/scion/go/border/qos/queues"
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
}

type surplus struct {
	Surplus  int
	Payments []int
}

var _ SchedulerInterface = (*MinMaxDeficitRoundRobinScheduler)(nil)

func (sched *MinMaxDeficitRoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	sched.schedulerSurplusMtx = &sync.Mutex{}
	sched.schedulerSurplus = surplus{0, make([]int, sched.totalLength)}

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}

}

func (sched *MinMaxDeficitRoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		for i := 0; i < sched.totalLength; i++ {
			sched.dequeue(routerConfig, forwarder, i)
		}
	}
}

func (sched *MinMaxDeficitRoundRobinScheduler) dequeue(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := routerConfig.Queues[queueNo].GetLength()
	pktToDequeue := min(64*(routerConfig.Queues[queueNo].GetMinBandwidth()/sched.quantumSum), 1)

	log.Debug("The queue has length", "length", length)
	log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	if length > 0 {

		if sched.surplusAvailable() {
			log.Debug("Surplus available", "surplus", sched.schedulerSurplus)
			if length > pktToDequeue {
				pktToDequeue = sched.getFromSurplus(routerConfig, queueNo, length)
				log.Debug("Dequeueing above minimum", "quantum", pktToDequeue)
			} else {
				if pktToDequeue-length > 0 {
					sched.payIntoSurplus(routerConfig, queueNo, pktToDequeue-length)
					log.Debug("Paying into surplus", "payment", pktToDequeue-length)
				}
			}
		}

		qps := routerConfig.Queues[queueNo].PopMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			forwarder(qp.Rp)
		}
	}
}

func (sched *MinMaxDeficitRoundRobinScheduler) getFromSurplus(routerConfig queues.InternalRouterConfig, queueNo int, request int) int {

	sched.schedulerSurplusMtx.Lock()
	defer sched.schedulerSurplusMtx.Unlock()

	// Check limit for queue
	// Take out of surplus

	upperLimit := min(64*(routerConfig.Queues[queueNo].GetMinBandwidth()/sched.quantumSum), 1)

	credit := min(sched.schedulerSurplus.Surplus, upperLimit)

	sched.schedulerSurplus.Surplus -= credit

	return credit

}

func (sched *MinMaxDeficitRoundRobinScheduler) payIntoSurplus(routerConfig queues.InternalRouterConfig, queueNo int, payment int) {

	sched.schedulerSurplusMtx.Lock()
	defer sched.schedulerSurplusMtx.Unlock()

	sched.schedulerSurplus.Surplus = min(sched.schedulerSurplus.Surplus+(payment-sched.schedulerSurplus.Payments[queueNo]), 0)
	sched.schedulerSurplus.Payments[queueNo] = payment

}

func (sched *MinMaxDeficitRoundRobinScheduler) surplusAvailable() bool {

	sched.schedulerSurplusMtx.Lock()
	defer sched.schedulerSurplusMtx.Unlock()

	return sched.schedulerSurplus.Surplus > 0
}

func (sched *MinMaxDeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
