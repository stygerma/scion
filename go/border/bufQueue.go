package main

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"

	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/ringbuf"
)

type packetBufQueue struct {
	Name         string          `yaml:"name"`
	ID           int             `yaml:"id"`
	MinBandwidth int             `yaml:"CIR"`
	MaxBandWidth int             `yaml:"PIR"`
	PoliceRate   int             `yaml:"policeRate"`
	MaxLength    int             `yaml:"maxLength"`
	priority     int             `yaml:"priority"`
	Profile      []actionProfile `yaml:"profile"`

	mutex *sync.Mutex

	bufQueue *ringbuf.Ring
	length   int
	tb       tokenBucket
}

type qPktList []qPkt
type qPktPointerList []*qPkt

var pkts = make(ringbuf.EntryList, 1)

func (pq *packetBufQueue) initQueue(mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{
		MaxBandWidth: pq.PoliceRate,
		tokens:       pq.PoliceRate,
		lastRefill:   time.Now(),
		mutex:        mutTb}
	pq.bufQueue = ringbuf.New(pq.MaxLength, nil, pq.Name)

}

func (pq *packetBufQueue) enqueue(rp *qPkt) {

	pq.bufQueue.Write(ringbuf.EntryList{rp}, false)

}

func (pq *packetBufQueue) canDequeue() bool {

	return pq.getLength() > 0
}

func (pq *packetBufQueue) getFillLevel() int {

	return pq.getLength() / pq.MaxLength
}

func (pq *packetBufQueue) getLength() int {

	return pq.bufQueue.Length()
}

func (pq *packetBufQueue) pop() *qPkt {

	// pkts := make(ringbuf.EntryList, 1)

	_, _ = pq.bufQueue.Read(pkts, false)

	return pkts[0].(*qPkt)
}

func (pq *packetBufQueue) popMultiple(number int) []*qPkt {

	// pkts := make(ringbuf.EntryList, number)

	_, _ = pq.bufQueue.Read(pkts, false)

	dubdub := (*[]*qPkt)(unsafe.Pointer(&pkts))

	retArr := *dubdub

	// retArr := make([]*qPkt, number)

	// for k, pkt := range pkts {
	// 	retArr[k] = pkt.(*qPkt)
	// }

	return retArr
}

func (pq *packetBufQueue) checkAction() policeAction {

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

func (pq *packetBufQueue) police(qp *qPkt, shouldLog bool) policeAction {

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

func (pq *packetBufQueue) getMinBandwidth() int {
	return pq.MinBandwidth
}

func (pq *packetBufQueue) getPriority() int {
	return pq.priority
}
