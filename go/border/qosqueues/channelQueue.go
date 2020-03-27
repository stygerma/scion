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
	"sync/atomic"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type ChannelPacketQueue struct {
	pktQue PacketQueue

	mutex *sync.Mutex

	queue  chan *QPkt
	length uint64
	tb     tokenBucket
	head   int
	tail   int
	mask   int
}

var _ PacketQueueInterface = (*ChannelPacketQueue)(nil)

func (pq *ChannelPacketQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.pktQue = que
	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.pktQue.PoliceRate,
		tokens:       pq.pktQue.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}
	pq.queue = make(chan *QPkt, pq.pktQue.MaxLength)
	pq.head = 0
	pq.tail = 0
	pq.mask = pq.pktQue.MaxLength - 1
}

func (pq *ChannelPacketQueue) Enqueue(rp *QPkt) {

	pq.queue <- rp

	atomic.AddUint64(&pq.length, 1)

}

func (pq *ChannelPacketQueue) canEnqueue() bool {

	return int(pq.length) < pq.pktQue.MaxLength
}

func (pq *ChannelPacketQueue) canDequeue() bool {

	return pq.head < pq.tail
}

func (pq *ChannelPacketQueue) GetFillLevel() int {

	return int(pq.length) / int(pq.pktQue.MaxLength)
}

func (pq *ChannelPacketQueue) GetLength() int {

	return int(pq.length)
}

func (pq *ChannelPacketQueue) peek() *QPkt {

	return nil
}

func (pq *ChannelPacketQueue) Pop() *QPkt {

	c := 1
	atomic.AddUint64(&pq.length, ^uint64(c-1))

	return <-pq.queue
}

func (pq *ChannelPacketQueue) PopMultiple(number int) []*QPkt {

	c := number
	atomic.AddUint64(&pq.length, ^uint64(c-1))

	pkts := make([]*QPkt, number)

	for i := 0; i < number; i++ {
		pkts[i] = <-pq.queue
	}

	return pkts
}

func (pq *ChannelPacketQueue) CheckAction() PoliceAction {

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

func (pq *ChannelPacketQueue) Police(qp *QPkt) PoliceAction {
	pq.tb.mutex.Lock()
	defer pq.tb.mutex.Unlock()

	packetSize := (qp.Rp.Bytes().Len()) // In byte

	tokenForPacket := packetSize * 8 // In bit

	log.Trace("Overall available bandwidth per second", "MaxBandWidth", pq.tb.MaxBandWidth)
	log.Trace("Spent token in last period", "#tokens", pq.tb.tokenSpent)
	log.Trace("Available bandwidth before refill", "bandwidth", pq.tb.tokens)

	pq.tb.refill()

	log.Trace("Available bandwidth after refill", "bandwidth", pq.tb.tokens)
	log.Trace("Tokens necessary for packet", "tokens", tokenForPacket)
	log.Trace("Tokens necessary for packet", "bytes", qp.Rp.Bytes().Len())

	if pq.tb.tokens-tokenForPacket > 0 {
		pq.tb.tokens = pq.tb.tokens - tokenForPacket
		pq.tb.tokenSpent += tokenForPacket
		qp.Act.action = PASS
		qp.Act.reason = None
	} else {
		qp.Act.action = DROP
		qp.Act.reason = BandWidthExceeded
	}

	log.Trace("Available bandwidth after update", "bandwidth", pq.tb.tokens)

	return qp.Act.action
}

func (pq *ChannelPacketQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *ChannelPacketQueue) GetPriority() int {
	return pq.pktQue.Priority
}

func (pq *ChannelPacketQueue) GetPacketQueue() PacketQueue {
	return pq.pktQue
}
