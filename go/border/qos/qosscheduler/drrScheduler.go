package qosscheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

type DeficitRoundRobinScheduler struct {
	quantumSum  int
	totalLength int
	messages    chan bool
}

var _ SchedulerInterface = (*DeficitRoundRobinScheduler)(nil)

func (sched *DeficitRoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	for i := 0; i < sched.totalLength; i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}

}

func getNoPacketsToDequeue(totalLength int, priority int, totalPriority int) int {
	return int(100.0 / float64(totalPriority) * float64(priority))
}

func (sched *DeficitRoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	// length := queue.GetLength()
	// nopkts := 64 * (queue.GetPriority() / sched.quantumSum)
	nopkts := getNoPacketsToDequeue(-1, queue.GetPriority(), sched.quantumSum)
	pktToDequeue := min(1, nopkts)

	var qp *qosqueues.QPkt

	for i := 0; i < pktToDequeue; i++ {
		qp = queue.Pop()
		if qp == nil {
			break
		}
		forwarder(qp.Rp)
	}
}

func (sched *DeficitRoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		time.Sleep(500 * time.Nanosecond)
	}
}

func (sched *DeficitRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
