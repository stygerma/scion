// Copyright 2020 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// RoundRobinScheduler is a standard round robin dequeue ignoring things like priority
type RoundRobinScheduler struct {
	totalLength   int
	messages      chan bool
	sleepDuration int
	tb            qosqueues.TokenBucket
}

var _ SchedulerInterface = (*RoundRobinScheduler)(nil)

func (sched *RoundRobinScheduler) Init(routerConfig queues.InternalRouterConfig) {
	sched.totalLength = len(routerConfig.Queues)
	sched.messages = make(chan bool)

	sched.tb.Init(routerConfig.Scheduler.Bandwidth)
	sched.sleepDuration = routerConfig.Scheduler.Latency
}

func (sched *RoundRobinScheduler) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	length := queue.GetLength()
	var qp *qosqueues.QPkt

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

func (sched *RoundRobinScheduler) Dequeuer(routerConfig queues.InternalRouterConfig,
	forwarder func(rp *rpkt.RtrPkt)) {

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
