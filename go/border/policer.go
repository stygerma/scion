package main

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type tokenBucket struct {
	MaxBandWidth     int // In bps
	tokens           int // One token is 1 b
	timerGranularity int
	lastRefill       time.Time
	mutex            *sync.Mutex
}

func (tb *tokenBucket) start() {

	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.tokens = tb.MaxBandWidth
	tb.lastRefill = time.Now()

	// timer1 := time.NewTimer(10 * time.Millisecond)

	// go func() {
	// 	for {
	// 		<-timer1.C
	// 		// TODO for some reason I can't import math and can't use the max function. Maybe this resolves itself magically in the future
	// 		tb.mutex.Lock()
	// 		if(tb.tokens+(tb.MaxBandWidth/100) > tb.MaxBandWidth) {
	// 			tb.tokens = tb.MaxBandWidth
	// 		} else {
	// 			tb.tokens = tb.tokens+(tb.MaxBandWidth/100)
	// 		}
	// 		tb.mutex.Unlock()
	// 	}
	// }()
}

func (tb * tokenBucket) refill() {

	now := time.Now()

	timeSinceLastUpdate := now.Sub(tb.lastRefill).Milliseconds()

	log.Debug("Last update was ", "Update time", timeSinceLastUpdate)

	if timeSinceLastUpdate > 100 {

		newTokens := ((tb.MaxBandWidth) * int(timeSinceLastUpdate)) / (1000 * 10)
		tb.lastRefill = now

		log.Debug("Add new tokens ", "#tokens", newTokens)

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

	packetSize := (qp.rp.Bytes().Len()) // In b

	tokenForPacket := packetSize // In b

	log.Debug("Available bandwidth before refill", "bandwidth", pq.tb.tokens)

	pq.tb.refill()

	log.Debug("Available bandwidth after refill", "bandwidth", pq.tb.tokens)
	log.Debug("Tokens necessary for packet", "tokens", tokenForPacket)
	log.Debug("Tokens necessary for packet", "bytes", qp.rp.Bytes().Len())

	if pq.tb.tokens - tokenForPacket > 0 {
		pq.tb.tokens = pq.tb.tokens - tokenForPacket
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
