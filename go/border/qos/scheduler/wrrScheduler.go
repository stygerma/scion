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

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is a deficit round robin dequeuer.
// Queues with higher priority will have more packets dequeued at the same time.

type WeightedRoundRobinScheduler struct {
	quantumSum       int
	totalLength      int
	messages         chan bool
	totalQueueLength int
	sleepDuration    int
	tb               queues.TokenBucket
	logger           ScheduleLogger
}

var _ SchedulerInterface = (*WeightedRoundRobinScheduler)(nil)

func (sched *WeightedRoundRobinScheduler) Init(routerConfig *queues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)
	var messageLen int
	for i := 0; i < len(routerConfig.Queues); i++ {
		messageLen += routerConfig.Queues[i].GetCapacity()
	}

	sched.messages = make(chan bool, messageLen)

	sched.logger = initLogger(sched.totalLength)

	for i := 0; i < len(routerConfig.Queues); i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}
	sched.tb.Init(routerConfig.Scheduler.Bandwidth)
	sched.sleepDuration = routerConfig.Scheduler.Latency
}

func getNoPacketsToDequeue(totalLength int, priority int, totalPriority int) int {
	return priority
}

func (sched *WeightedRoundRobinScheduler) Dequeue(
	queue queues.PacketQueueInterface,
	forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	var qp *queues.QPkt

	pktToDequeue := getNoPacketsToDequeue(sched.totalQueueLength, queue.GetPriority(), sched.quantumSum)

	sched.logger.attempted[queueNo] += pktToDequeue

	for i := 0; i < pktToDequeue; i++ {

		qp = queue.Pop()

		if qp == nil {
			break
		}

		pktLen = len(qp.Rp.Raw)

		amount0 += pktLen

		for !(sched.tb.Take(pktLen)) {
			time.Sleep(30 * time.Millisecond)
		}

		sched.logger.lastRound[queueNo]++
		sched.logger.total[queueNo]++

		qp.Mtx.Lock()
		if (uint8(qp.Act.GetAction()) == 1) && !qp.Forward { //TODO: find smarter way uint8(qp.Act.GetAction()) == 0 ||
			// if !qp.Forward {
			qp.Forward = true
			qp.Mtx.Unlock()
			log.Debug("Packet in weightedRoundRobinScheduler forwarding enabled", "forwardBool", qp.Forward, "id", qp.Rp.Id)

			return
		}
		qp.Mtx.Unlock()
		log.Debug("Packet in weightedRoundRobinScheduler forwarded", "id", qp.Rp.Id)

		forwarder(qp.Rp)

	}
}

func (sched *WeightedRoundRobinScheduler) Dequeuer(routerConfig *queues.InternalRouterConfig,
	forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	sleepDuration := time.Duration(time.Duration(sched.sleepDuration) * time.Microsecond)
	for <-sched.messages {
		t0 := time.Now()
		sched.totalQueueLength = 0
		for i := 0; i < sched.totalLength; i++ {
			sched.totalQueueLength += routerConfig.Queues[i].GetLength()
		}

		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}

		sched.showLog(*routerConfig)

		for time.Now().Sub(t0) < sleepDuration {
			time.Sleep(time.Duration(sched.sleepDuration/10) * time.Microsecond)
		}
	}
}

var amount0 int

func (sched *WeightedRoundRobinScheduler) UpdateIncoming(queueNo int) {
	sched.logger.incoming[queueNo]++
}

func (sched *WeightedRoundRobinScheduler) showLog(routerConfig queues.InternalRouterConfig) {

	sched.logger.iterations++
	if time.Now().Sub(sched.logger.t0) > time.Duration(1*time.Second) {

		var queLen = make([]int, sched.totalLength)
		for i := 0; i < sched.totalLength; i++ {
			queLen[i] = routerConfig.Queues[i].GetLength()
		}
		// log.Debug("STAT",
		// 	"iterations", sched.logger.iterations,
		// 	"incoming", sched.logger.incoming,
		// 	"deqLastRound",
		// 	sched.logger.lastRound, "deqAttempted",
		// 	sched.logger.attempted, "deqTotal",
		// 	sched.logger.total, "currQueueLen", queLen)
		for i := 0; i < len(sched.logger.lastRound); i++ {
			sched.logger.lastRound[i] = 0
		}
		for i := 0; i < len(sched.logger.attempted); i++ {

			sched.logger.attempted[i] = 0
		}
		for i := 0; i < len(sched.logger.incoming); i++ {
			sched.logger.incoming[i] = 0
		}
		sched.logger.t0 = time.Now()
		sched.logger.iterations = 0
	}

}

func (sched *WeightedRoundRobinScheduler) GetMessages() *chan bool {
	return &sched.messages
}
