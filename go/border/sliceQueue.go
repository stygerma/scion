package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type packetSliceQueue struct {
	Name         string          `yaml:"name"`
	ID           int             `yaml:"id"`
	MinBandwidth int             `yaml:"CIR"`
	MaxBandWidth int             `yaml:"PIR"`
	PoliceRate   int             `yaml:"policeRate"`
	MaxLength    int             `yaml:"maxLength"`
	priority     int             `yaml:"priority"`
	Profile      []actionProfile `yaml:"profile"`

	mutex *sync.Mutex

	queue  []*qPkt
	length int
	tb     tokenBucket
}

func (pq *packetSliceQueue) initQueue(mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.PoliceRate,
		tokens:       pq.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}

}

func (pq *packetSliceQueue) enqueue(rp *qPkt) {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queue = append(pq.queue, rp)
	pq.length = pq.length + 1

}

func (pq *packetSliceQueue) canDequeue() bool {

	return pq.length > 0
}

func (pq *packetSliceQueue) getFillLevel() int {

	return pq.length / pq.MaxLength
}

func (pq *packetSliceQueue) getLength() int {

	return pq.length
}

func (pq *packetSliceQueue) peek() *qPkt {

	return pq.queue[0]
}

func (pq *packetSliceQueue) pop() *qPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[0]
	pq.queue = pq.queue[1:]
	pq.length = pq.length - 1

	return pkt
}

func (pq *packetSliceQueue) popMultiple(number int) []*qPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[:number]
	pq.queue = pq.queue[number:]
	pq.length = pq.length - number

	return pkt
}

func (pq *packetSliceQueue) checkAction() policeAction {

	level := pq.getFillLevel()

	log.Info("Current level is", "level", level)
	log.Info("Profiles are", "profiles", pq.Profile)

	for j := len(pq.Profile) - 1; j >= 0; j-- {
		if level >= pq.Profile[j].FillLevel {
			log.Info("Matched a rule!")
			if rand.Intn(100) < (pq.Profile[j].Prob) {
				log.Info("Take Action!")
				return pq.Profile[j].Action
			}
			log.Info("Do not take Action")

		}
	}

	return PASS
}

func (pq *packetSliceQueue) police(qp *qPkt, shouldLog bool) policeAction {
	pq.tb.mutex.Lock()
	defer pq.tb.mutex.Unlock()

	packetSize := (qp.rp.Bytes().Len()) // In byte

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
		log.Debug("Tokens necessary for packet", "bytes", qp.rp.Bytes().Len())
	}

	if pq.tb.tokens-tokenForPacket > 0 {
		pq.tb.tokens = pq.tb.tokens - tokenForPacket
		pq.tb.tokenSpent += tokenForPacket
		qp.act.action = PASS
		qp.act.reason = None
	} else {
		qp.act.action = DROP
		qp.act.reason = BandWidthExceeded
	}

	if shouldLog {
		log.Debug("Available bandwidth after update", "bandwidth", pq.tb.tokens)
	}

	return qp.act.action
}

func (pq *packetSliceQueue) getMinBandwidth() int {
	return pq.MinBandwidth
}

func (pq *packetSliceQueue) getPriority() int {
	return pq.priority
}
