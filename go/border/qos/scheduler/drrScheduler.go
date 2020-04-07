package scheduler

import (
	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

type deficitRoundRobinScheduler struct {
	quantumSum  int
	totalLength int
	messages    chan bool
}

var _ SchedulerInterface = (*deficitRoundRobinScheduler)(nil)

func (sched *deficitRoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}

}

func (sched *deficitRoundRobinScheduler) dequeue(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := routerConfig.Queues[queueNo].GetLength()
	var nopkts int = 64 * (routerConfig.Queues[queueNo].GetPriority() / sched.quantumSum)
	pktToDequeue := min(1, nopkts)

	log.Debug("The queue has length", "length", length)
	log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	if length > 0 {
		qps := routerConfig.Queues[queueNo].PopMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			forwarder(qp.Rp)
		}
	}
}

func (sched *deficitRoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		for i := 0; i < sched.totalLength; i++ {
			sched.dequeue(routerConfig, forwarder, i)
		}
	}
}

func (sched *deficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}