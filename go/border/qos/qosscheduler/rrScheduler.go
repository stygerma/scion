package qosscheduler

import (
	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
)

type RoundRobinScheduler struct {
	totalLength   int
	messages      chan bool
	sleepTime     int
	sleptLastTime bool
}

var _ SchedulerInterface = (*RoundRobinScheduler)(nil)

// This is a standard round robin dequeue ignoring things like priority

func (sched *RoundRobinScheduler) Init(routerConfig qosqueues.InternalRouterConfig) {
	sched.totalLength = len(routerConfig.Queues)
	sched.messages = make(chan bool)
	sched.sleepTime = 2
	sched.sleptLastTime = true
}

func (sched *RoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := queue.GetLength()
	var qp *qosqueues.QPkt

	for i := 0; i < length; i++ {
		qp = queue.Pop()
		forwarder(qp.Rp)
	}

	// if length > 0 {
	// qps := queue.PopMultiple(length)
	// for _, qp := range qps {
	// 	forwarder(qp.Rp)
	// }

	// }
	// log.Debug("Finished Dequeue")
}

func (sched *RoundRobinScheduler) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		// log.Debug("Start of Dequeuer")
		// select {
		// case <-sched.messages:
		// 	// sched.sleptLastTime = false
		// default:
		// 	// if sched.sleptLastTime {
		// 	// 	sched.sleepTime = max(sched.sleepTime*2, 2)
		// 	// } else {
		// 	// 	sched.sleepTime = 2
		// 	// }
		// 	// sched.sleptLastTime = true
		// 	// time.Sleep(1 * time.Millisecond)
		// }
		// time.Sleep(10 * time.Millisecond)
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
	}
}

func (sched *RoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}

// func max(a, b int) int {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }
