package qosqueues

import (
	"math/rand"
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type packetSliceQueue struct {
	pktQue PacketQueue

	mutex *sync.Mutex

	queue  []*QPkt
	length int
	tb     tokenBucket
}

var _ PacketQueueInterface = (*packetSliceQueue)(nil)

func (pq *packetSliceQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.pktQue = que
	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.pktQue.PoliceRate,
		tokens:       pq.pktQue.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}

}

func (pq *packetSliceQueue) Enqueue(rp *QPkt) {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queue = append(pq.queue, rp)
	pq.length = pq.length + 1

}

func (pq *packetSliceQueue) canDequeue() bool {

	return pq.length > 0
}

func (pq *packetSliceQueue) GetFillLevel() int {

	return pq.length / pq.pktQue.MaxLength
}

func (pq *packetSliceQueue) GetLength() int {

	return pq.length
}

func (pq *packetSliceQueue) peek() *QPkt {

	return pq.queue[0]
}

func (pq *packetSliceQueue) Pop() *QPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[0]
	pq.queue = pq.queue[1:]
	pq.length = pq.length - 1

	return pkt
}

func (pq *packetSliceQueue) PopMultiple(number int) []*QPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[:number]
	pq.queue = pq.queue[number:]
	pq.length = pq.length - number

	return pkt
}

func (pq *packetSliceQueue) CheckAction() PoliceAction {

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

func (pq *packetSliceQueue) Police(qp *QPkt, shouldLog bool) PoliceAction {
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

func (pq *packetSliceQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *packetSliceQueue) GetPriority() int {
	return pq.pktQue.Priority
}
