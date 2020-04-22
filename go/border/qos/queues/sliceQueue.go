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

package queues

import (
	"math/rand"
	"sync"

	"github.com/scionproto/scion/go/border/qos/conf"
)

type PacketSliceQueue struct {
	pktQue PacketQueue
	mutex  *sync.Mutex
	queue  []*QPkt
	tb     tokenBucket
}

var _ PacketQueueInterface = (*PacketSliceQueue)(nil)

func (pq *PacketSliceQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {
	pq.pktQue = que
	pq.mutex = mutQue
	pq.queue = make([]*QPkt, 0)
	pq.tb = tokenBucket{}
	pq.tb.Init(pq.pktQue.PoliceRate)
}

func (pq *PacketSliceQueue) Enqueue(rp *QPkt) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queue = append(pq.queue, rp)
}

func (pq *PacketSliceQueue) canDequeue() bool {
	return len(pq.queue) > 0
}

func (pq *PacketSliceQueue) GetFillLevel() int {
	return len(pq.queue) / pq.pktQue.MaxLength
}

func (pq *PacketSliceQueue) GetLength() int {
	return len(pq.queue)
}

func (pq *PacketSliceQueue) peek() *QPkt {
	return pq.queue[0]
}

func (pq *PacketSliceQueue) Pop() *QPkt {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[0]
	pq.queue = pq.queue[1:]
	return pkt
}

func (pq *PacketSliceQueue) PopMultiple(number int) []*QPkt {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[:number]
	pq.queue = pq.queue[number:]
	return pkt
}

func (pq *PacketSliceQueue) CheckAction() conf.PoliceAction {
	level := pq.GetFillLevel()
	for j := len(pq.pktQue.Profile) - 1; j >= 0; j-- {
		if level >= pq.pktQue.Profile[j].FillLevel {
<<<<<<< 82102bf27a694b6125fa87ce957aea70d5cc5377
			//log.Trace("Matched a rule!")
			rand := rand.Intn(100)
			if rand < (pq.pktQue.Profile[j].Prob) {
				//log.Trace("Take Action!")
=======
			if rand.Intn(100) < (pq.pktQue.Profile[j].Prob) {
>>>>>>> cleanup
				return pq.pktQue.Profile[j].Action
			}
		}
	}
	return conf.PASS
}

func (pq *PacketSliceQueue) Police(qp *QPkt) conf.PoliceAction {
	return pq.tb.PoliceBucket(qp)
}

func (pq *PacketSliceQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *PacketSliceQueue) GetMaxBandwidth() int {
	return pq.pktQue.MaxBandWidth
}

func (pq *PacketSliceQueue) GetPriority() int {
	return pq.pktQue.Priority
}

func (pq *PacketSliceQueue) GetPacketQueue() PacketQueue {
	return pq.pktQue
}
