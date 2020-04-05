package qosscheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

type deficitRoundRobinScheduler struct {
	quantumSum  int
	totalLength int
	messages    chan bool
}

var _ SchedulerInterface = (*deficitRoundRobinScheduler)(nil)

func (sched *deficitRoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}

}

func (sched *deficitRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := queue.GetLength()
	var nopkts int = 64 * (queue.GetPriority() / sched.quantumSum)
	pktToDequeue := min(1, nopkts)

	if length > 0 {
		qps := queue.PopMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			forwarder(qp.Rp)
		}
	}
}

func (sched *deficitRoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (sched *deficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
