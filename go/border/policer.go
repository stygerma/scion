package main

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type tokenBucket struct {
	MaxBandWidth     int // In bps
	tokens           int // One token is 1 b
	tokenSpent		 int
	timerGranularity int
	lastRefill       time.Time
	mutex            *sync.Mutex
}

func (tb * tokenBucket) refill() {

	now := time.Now()

	timeSinceLastUpdate := now.Sub(tb.lastRefill).Milliseconds()

	log.Debug("Last update was ", "ms ago", timeSinceLastUpdate)

	if timeSinceLastUpdate > 100 {

		newTokens := ((tb.MaxBandWidth) * int(timeSinceLastUpdate)) / (1000)
		tb.lastRefill = now

		log.Debug("Add new tokens ", "#tokens", newTokens)
		log.Debug("Spent token in last period ", "#tokens", tb.tokenSpent)
		tb.tokenSpent = 0

		if tb.tokens + newTokens > tb.MaxBandWidth {
			tb.tokens = tb.MaxBandWidth
		} else {
			tb.tokens = tb.tokens + newTokens
		}
	}

}

func (pq *packetQueue) police(qp *qPkt) policeAction {
	pq.tb.mutex.Lock()
	defer pq.tb.mutex.Unlock()

	packetSize := (qp.rp.Bytes().Len()) // In byte

	tokenForPacket := packetSize * 8 // In bit

	log.Debug("Available bandwidth before refill", "bandwidth", pq.tb.tokens)

	pq.tb.refill()

	log.Debug("Available bandwidth after refill", "bandwidth", pq.tb.tokens)
	log.Debug("Tokens necessary for packet", "tokens", tokenForPacket)
	log.Debug("Tokens necessary for packet", "bytes", qp.rp.Bytes().Len())

	if pq.tb.tokens - tokenForPacket > 0 {
		pq.tb.tokens = pq.tb.tokens - tokenForPacket
		pq.tb.tokenSpent += tokenForPacket
	} else {
		qp.act.action = DROP
		qp.act.reason = BandWidthExceeded
	}

	log.Debug("Available bandwidth after update", "bandwidth", pq.tb.tokens)

	return qp.act.action
}

func (qp *qPkt) sendNotification() {
	select {
	case r.notifications <- qp:
	default:
	}
}
