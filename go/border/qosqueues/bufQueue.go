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
	"github.com/scionproto/scion/go/lib/ringbuf"
)

type PacketBufQueue struct {
	pktQue PacketQueue

	mutex *sync.Mutex

	bufQueue *ringbuf.Ring
	length   int
	tb       tokenBucket
}

// type QPktList []QPkt
// type QPktPointerList []*QPkt

var _ PacketQueueInterface = (*PacketBufQueue)(nil)

func (pq *PacketBufQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.pktQue = que
	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.pktQue.PoliceRate,
		tokens:       pq.pktQue.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}
	pq.bufQueue = ringbuf.New(pq.pktQue.MaxLength, func() interface{} {
		return &QPkt{}
	}, pq.pktQue.Name)

}

func (pq *PacketBufQueue) Enqueue(rp *QPkt) {

	pq.bufQueue.Write(ringbuf.EntryList{rp}, false)

}

func (pq *PacketBufQueue) canDequeue() bool {

	return pq.GetLength() > 0
}

func (pq *PacketBufQueue) GetFillLevel() int {

	return pq.GetLength() / pq.pktQue.MaxLength
}

func (pq *PacketBufQueue) GetLength() int {

	return pq.bufQueue.Length()
}

func (pq *PacketBufQueue) Pop() *QPkt {

	pkts := make(ringbuf.EntryList, 1)

	_, _ = pq.bufQueue.Read(pkts, false)

	return pkts[0].(*QPkt)
}

func (pq *PacketBufQueue) PopMultiple(number int) []*QPkt {

	pkts := make(ringbuf.EntryList, number)

	_, _ = pq.bufQueue.Read(pkts, false)

	// dubdub := (*[]*QPkt)(unsafe.Pointer(&pkts))

	// retArr := *dubdub

	retArr := make([]*QPkt, number)

	for k, pkt := range pkts {
		retArr[k] = pkt.(*QPkt)
	}

	return retArr
}

func (pq *PacketBufQueue) CheckAction() PoliceAction {

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

func (pq *PacketBufQueue) Police(qp *QPkt) PoliceAction {
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
		pq.tb.tokens -= tokenForPacket
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

func (pq *PacketBufQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *PacketBufQueue) GetPriority() int {
	return pq.pktQue.Priority
}

func (pq *PacketBufQueue) GetPacketQueue() PacketQueue {
	return pq.pktQue
}
