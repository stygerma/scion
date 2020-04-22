package main

import (
	"math/rand"
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

func (pq *packetQueue) police(qp *qPkt, shouldLog bool) policeAction {
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

func (pq *packetQueue) checkAction() policeAction {

	level := pq.getFillLevel()

	log.Info("Current level is", "level", level)
	log.Info("Profiles are", "profiles", pq.Profile)

	for j := len(pq.Profile) - 1; j >= 0; j-- {
		if level >= pq.Profile[j].FillLevel {
			log.Info("Matched a rule!")
			if rand.Intn(100) < (pq.Profile[j].Prob) {
				log.Info("Take Action!")
				return pq.Profile[j].Action
			} else {
				log.Info("Do not take Action")
			}
		}
	}

	return PASS
}

func returnAction(polAction policeAction, profAction policeAction) policeAction {

	if polAction == DROPNOTIFY || profAction == DROPNOTIFY {
		return DROPNOTIFY
	}

	if polAction == DROP || profAction == DROP {
		return DROP
	}

	if polAction == NOTIFY || profAction == NOTIFY {
		return NOTIFY
	}

	return PASS
}

func (qp *qPkt) sendNotification() {
	select {
	case r.notifications <- qp:
	default:
	}
}
