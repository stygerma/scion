package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type customPacketQueue struct {
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
	head   int
	tail   int
	mask   int
}

func (pq *customPacketQueue) initQueue(mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.PoliceRate,
		tokens:       pq.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}
	pq.queue = make([]*qPkt, pq.MaxLength)
	pq.head = 0
	pq.tail = 0
	pq.mask = pq.MaxLength - 1

	// fmt.Println("Finish init")
}

func (pq *customPacketQueue) enqueue(rp *qPkt) {

	// TODO: Making this lockfree makes it 10 times faster
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Enqueue at", pq.tail, "Dequeue at", pq.head)
	pq.queue[pq.tail] = rp
	pq.tail = (pq.tail + 1) & pq.mask
	pq.length = pq.length + 1

}

func (pq *customPacketQueue) canEnqueue() bool {

	return pq.length < pq.MaxLength
}

func (pq *customPacketQueue) canDequeue() bool {

	return pq.head < pq.tail
}

func (pq *customPacketQueue) getFillLevel() int {

	return pq.length / pq.MaxLength
}

func (pq *customPacketQueue) getLength() int {

	return pq.length
}

func (pq *customPacketQueue) peek() *qPkt {

	return pq.queue[0]
}

func (pq *customPacketQueue) pop() *qPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Enqueue at", pq.tail, "Dequeue at", pq.head)

	pkt := pq.queue[pq.head]
	pq.head = (pq.head + pq.MaxLength + 1) & pq.mask
	pq.length = pq.length - 1

	return pkt
}

func (pq *customPacketQueue) popMultiple(number int) []*qPkt {

	// TODO: Readd this as soon as popMultiple works as standalone
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	// fmt.Println("Pop 10")

	var pkt []*qPkt

	if pq.head+number < pq.MaxLength {
		pkt = pq.queue[pq.head : pq.head+number]
		pq.head = (pq.head + number) & pq.mask

	} else {
		for pq.head+number > pq.MaxLength {
			pkt = pq.queue[pq.head:pq.MaxLength]
			number = number - (pq.MaxLength - pq.head)
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

func (pq *customPacketQueue) checkAction() policeAction {

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

func (pq *customPacketQueue) police(qp *qPkt, shouldLog bool) policeAction {
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

func (pq *customPacketQueue) getMinBandwidth() int {
	return pq.MinBandwidth
}

func (pq *customPacketQueue) getPriority() int {
	return pq.priority
}
