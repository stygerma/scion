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
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type customPacketQueue struct {
	pktQue PacketQueue

	mutex *sync.Mutex

	queue  []*QPkt
	length int
	tb     tokenBucket
	head   int
	tail   int
	mask   int
}

var _ PacketQueueInterface = (*customPacketQueue)(nil)

func (pq *customPacketQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.pktQue = que
	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.pktQue.PoliceRate,
		tokens:       pq.pktQue.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}
	pq.queue = make([]*QPkt, pq.pktQue.MaxLength)
	pq.head = 0
	pq.tail = 0
	pq.mask = pq.pktQue.MaxLength - 1

	// fmt.Println("Finish init")
}

func (pq *customPacketQueue) Enqueue(rp *QPkt) {

	// TODO: Making this lockfree makes it 10 times faster
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Enqueue at", pq.tail, "Dequeue at", pq.head)
	pq.queue[pq.tail] = rp
	pq.tail = (pq.tail + 1) & pq.mask
	pq.length = pq.length + 1

}

func (pq *customPacketQueue) canEnqueue() bool {

	return pq.length < pq.pktQue.MaxLength
}

func (pq *customPacketQueue) canDequeue() bool {

	return pq.head < pq.tail
}

func (pq *customPacketQueue) GetFillLevel() int {

	return pq.length / pq.pktQue.MaxLength
}

func (pq *customPacketQueue) GetLength() int {

	return pq.length
}

func (pq *customPacketQueue) peek() *QPkt {

	return pq.queue[0]
}

func (pq *customPacketQueue) Pop() *QPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Enqueue at", pq.tail, "Dequeue at", pq.head)

	pkt := pq.queue[pq.head]
	pq.head = (pq.head + pq.pktQue.MaxLength + 1) & pq.mask
	pq.length = pq.length - 1

	return pkt
}

func (pq *customPacketQueue) PopMultiple(number int) []*QPkt {

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

func (pq *customPacketQueue) CheckAction() PoliceAction {

	level := pq.GetFillLevel()

	log.Info("Current level is", "level", level)
	log.Info("Profiles are", "profiles", pq.pktQue.Profile)

	for j := len(pq.pktQue.Profile) - 1; j >= 0; j-- {
		if level >= pq.pktQue.Profile[j].FillLevel {
			log.Info("Matched a rule!")
			if rand.Intn(100) < (pq.pktQue.Profile[j].Prob) {
				log.Info("Take Action!")
				return pq.pktQue.Profile[j].Action
			}
			log.Info("Do not take Action")

		}
	}

	return PASS
}

func (pq *customPacketQueue) Police(qp *QPkt, shouldLog bool) PoliceAction {
	pq.tb.mutex.Lock()
	defer pq.tb.mutex.Unlock()

	packetSize := (qp.Rp.Bytes().Len()) // In byte

	tokenForPacket := packetSize * 8 // In bit

	if shouldLog {
		log.Debug("Overall available bandwidth per second", "MaxBandWidth", pq.tb.MaxBandWidth)
		log.Debug("Spent token in last period", "#tokens", pq.tb.tokenSpent)
		log.Debug("Available bandwidth before refill", "bandwidth", pq.tb.tokens)
	}

	pq.tb.refill(shouldLog)

	if shouldLog {
		log.Debug("Available bandwidth after refill", "bandwidth", pq.tb.tokens)
		log.Debug("Tokens necessary for packet", "tokens", tokenForPacket)
		log.Debug("Tokens necessary for packet", "bytes", qp.Rp.Bytes().Len())
	}

	if pq.tb.tokens-tokenForPacket > 0 {
		pq.tb.tokens = pq.tb.tokens - tokenForPacket
		pq.tb.tokenSpent += tokenForPacket
		qp.Act.action = PASS
		qp.Act.reason = None
	} else {
		qp.Act.action = DROP
		qp.Act.reason = BandWidthExceeded
	}

	if shouldLog {
		log.Debug("Available bandwidth after update", "bandwidth", pq.tb.tokens)
	}

	return qp.Act.action
}

func (pq *customPacketQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *customPacketQueue) GetPriority() int {
	return pq.pktQue.Priority
}
