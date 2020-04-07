package scheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
)

type RoundRobinScheduler struct {
	totalLength   int
	messages      chan bool
	sleepDuration int
	tb            qosqueues.TokenBucket
}

var _ SchedulerInterface = (*RoundRobinScheduler)(nil)

// This is a standard round robin dequeue ignoring things like priority

func (sched *RoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {
	sched.totalLength = len(routerConfig.Queues)
	sched.messages = make(chan bool)

	sched.tb.Init(routerConfig.Scheduler.Bandwidth)
	sched.sleepDuration = routerConfig.Scheduler.Latency
}

<<<<<<< HEAD:go/border/qos/scheduler/rrScheduler.go
func (sched *RoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := queue.GetLength()
	var qp *qosqueues.QPkt
=======
func (sched *RoundRobinScheduler) dequeue(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt), queueNo int) {
>>>>>>> 00e2ea31c... refactor.:go/border/qos/scheduler/rrScheduler.go

	for i := 0; i < length; i++ {
		qp = queue.Pop()
		if qp == nil {
			continue
		}

		for !(sched.tb.Take(qp.Rp.Bytes().Len())) {
			time.Sleep(50 * time.Millisecond)
		}

		forwarder(qp.Rp)
	}
}

func (sched *RoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	for {
		t0 := time.Now()
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}
		for time.Now().Sub(t0) < time.Duration(time.Duration(sched.sleepDuration)*time.Microsecond) {
			time.Sleep(time.Duration(sched.sleepDuration/10) * time.Microsecond)
		}
	}
}

func (sched *RoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
