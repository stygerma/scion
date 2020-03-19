package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type channelPacketQueue struct {
	Name         string          `yaml:"name"`
	ID           int             `yaml:"id"`
	MinBandwidth int             `yaml:"CIR"`
	MaxBandWidth int             `yaml:"PIR"`
	PoliceRate   int             `yaml:"policeRate"`
	MaxLength    int             `yaml:"maxLength"`
	priority     int             `yaml:"priority"`
	Profile      []actionProfile `yaml:"profile"`

	mutex *sync.Mutex

	queue  chan *qPkt
	length uint64
	tb     tokenBucket
	head   int
	tail   int
	mask   int
}

func (pq *channelPacketQueue) initQueue(mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.PoliceRate,
		tokens:       pq.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}
	pq.queue = make(chan *qPkt, pq.MaxLength)
	pq.head = 0
	pq.tail = 0
	pq.mask = pq.MaxLength - 1

	// fmt.Println("Finish init")
}

func (pq *channelPacketQueue) enqueue(rp *qPkt) {

	pq.queue <- rp

	atomic.AddUint64(&pq.length, 1)

}

func (pq *channelPacketQueue) canEnqueue() bool {

	return int(pq.length) < pq.MaxLength
}

func (pq *channelPacketQueue) canDequeue() bool {

	return pq.head < pq.tail
}

func (pq *channelPacketQueue) getFillLevel() int {

	return int(pq.length) / int(pq.MaxLength)
}

func (pq *channelPacketQueue) getLength() int {

	return int(pq.length)
}

func (pq *channelPacketQueue) peek() *qPkt {

	return nil
}

func (pq *channelPacketQueue) pop() *qPkt {

	c := 1
	atomic.AddUint64(&pq.length, ^uint64(c-1))

	return <-pq.queue
}

func (pq *channelPacketQueue) popMultiple(number int) []*qPkt {

	c := number
	atomic.AddUint64(&pq.length, ^uint64(c-1))

	pkts := make([]*qPkt, number)

	for i := 0; i < number; i++ {
		pkts[i] = <-pq.queue
	}

	return pkts
}

func (pq *channelPacketQueue) checkAction() policeAction {

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

func (pq *channelPacketQueue) police(qp *qPkt, shouldLog bool) policeAction {
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

func (pq *channelPacketQueue) getMinBandwidth() int {
	return pq.MinBandwidth
}

func (pq *channelPacketQueue) getPriority() int {
	return pq.priority
}
