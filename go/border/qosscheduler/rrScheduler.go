package qosscheduler

import (
	"github.com/scionproto/scion/go/border/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

type RoundRobinScheduler struct {
	totalLength int
}

var _ SchedulerInterface = (*RoundRobinScheduler)(nil)

// This is a standard round robin dequeue ignoring things like priority

func (sched *RoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {
	sched.totalLength = len(routerConfig.Queues)
}

func (sched *RoundRobinScheduler) dequeue(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := routerConfig.Queues[queueNo].GetLength()
	log.Debug("The queue has length", "length", length)

	if length > 0 {
		qps := routerConfig.Queues[queueNo].PopMultiple(length)
		for _, qp := range qps {
			forwarder(qp.Rp)
		}
	}
}

func (sched *RoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	for {
		for i := 0; i < sched.totalLength; i++ {
			sched.dequeue(routerConfig, forwarder, i)
		}
	}
}
