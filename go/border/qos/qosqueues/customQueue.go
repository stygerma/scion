// Copyright 2020 ETH Zurich
// Copyright 2020 ETH Zurich, Anapaya Systems
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

package qosqueues

import (
	"math/rand"
	"sync"
)

type CustomPacketQueue struct {
	pktQue PacketQueue

	mutex *sync.Mutex

	queue  []*QPkt
	length int
	tb     tokenBucket
	head   int
	tail   int
	mask   int
}

var _ PacketQueueInterface = (*CustomPacketQueue)(nil)

func (pq *CustomPacketQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.pktQue = que
	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{}
	pq.tb.Init(pq.pktQue.PoliceRate)
	pq.queue = make([]*QPkt, pq.pktQue.MaxLength)
	pq.head = 0
	pq.tail = 0
	pq.mask = pq.pktQue.MaxLength - 1

	// fmt.Println("Finish init")
}

func (pq *CustomPacketQueue) Enqueue(rp *QPkt) {

	// TODO: Making this lockfree makes it 10 times faster
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Enqueue at", pq.tail, "Dequeue at", pq.head)
	pq.queue[pq.tail] = rp
	pq.tail = (pq.tail + 1) & pq.mask
	pq.length = pq.length + 1

}

func (pq *CustomPacketQueue) canEnqueue() bool {

	return pq.length < pq.pktQue.MaxLength
}

func (pq *CustomPacketQueue) canDequeue() bool {

	return pq.head < pq.tail
}

func (pq *CustomPacketQueue) GetFillLevel() int {

	return pq.length / pq.pktQue.MaxLength
}

func (pq *CustomPacketQueue) GetLength() int {

	return pq.length
}

func (pq *CustomPacketQueue) peek() *QPkt {

	return pq.queue[0]
}

func (pq *CustomPacketQueue) Pop() *QPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Enqueue at", pq.tail, "Dequeue at", pq.head)

	pkt := pq.queue[pq.head]
	pq.head = (pq.head + pq.pktQue.MaxLength + 1) & pq.mask
	pq.length = pq.length - 1

	return pkt
}

func (pq *CustomPacketQueue) PopMultiple(number int) []*QPkt {

	// TODO: Readd this as soon as popMultiple works as standalone
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Pop 10")

	var pkt []*QPkt

	if pq.head+number < pq.pktQue.MaxLength {
		pkt = pq.queue[pq.head : pq.head+number]
		pq.head = (pq.head + number) & pq.mask

	} else {
		for pq.head+number > pq.pktQue.MaxLength {
			pkt = pq.queue[pq.head:pq.pktQue.MaxLength]
			number = number - (pq.pktQue.MaxLength - pq.head)
			pq.head = 0

			pkt = append(pkt, pq.queue[pq.head:pq.head+number]...)
			pq.head = (pq.head + number) & pq.mask
		}

		pkt = append(pkt, pq.queue[pq.head:pq.head+number]...)
		pq.head = (pq.head + number) & pq.mask

	}

	pq.length = pq.length - number

	return pkt
}

func (pq *CustomPacketQueue) CheckAction() PoliceAction {

	level := pq.GetFillLevel()

	//log.Trace("Current level is", "level", level)
	//log.Trace("Profiles are", "profiles", pq.pktQue.Profile)

	for j := len(pq.pktQue.Profile) - 1; j >= 0; j-- {
		if level >= pq.pktQue.Profile[j].FillLevel {
			//log.Trace("Matched a rule!")
			rand := rand.Intn(100)
			if rand < (pq.pktQue.Profile[j].Prob) {
				//log.Trace("Take Action!")
				return pq.pktQue.Profile[j].Action
			}
			//log.Trace("Do not take Action")

		}
	}

	return PASS
}

func (pq *CustomPacketQueue) Police(qp *QPkt) PoliceAction {
	return pq.tb.PoliceBucket(qp)
}

func (pq *CustomPacketQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *CustomPacketQueue) GetMaxBandwidth() int {
	return pq.pktQue.MaxBandWidth
}

func (pq *CustomPacketQueue) GetPriority() int {
	return pq.pktQue.Priority
}

func (pq *CustomPacketQueue) GetPacketQueue() PacketQueue {
	return pq.pktQue
}
