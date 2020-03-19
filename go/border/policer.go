package main

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type tokenBucket struct {
	MaxBandWidth     int // In bps
	tokens           int // One token is 1 b
	tokenSpent       int
	timerGranularity int
	lastRefill       time.Time
	mutex            *sync.Mutex
}

func (tb *tokenBucket) refill(shouldLog bool) {

	now := time.Now()

	timeSinceLastUpdate := now.Sub(tb.lastRefill).Milliseconds()

	if shouldLog {
		log.Debug("Last update was", "ms ago", timeSinceLastUpdate)
	}

	if timeSinceLastUpdate > 100 {

		newTokens := ((tb.MaxBandWidth) * int(timeSinceLastUpdate)) / (1000)
		tb.lastRefill = now

		if shouldLog {
			log.Debug("Add new tokens", "#tokens", newTokens)
			log.Debug("On Update: Spent token in last period", "#tokens", tb.tokenSpent)
		}

		tb.tokenSpent = 0

		if tb.tokens+newTokens > tb.MaxBandWidth {
			tb.tokens = tb.MaxBandWidth
		} else {
			tb.tokens = tb.tokens + newTokens
		}
	}

}

func (qp *qPkt) sendNotification() {
	select {
	case r.notifications <- qp:
	default:
	}
}
