package main

import (
	"sync"
	"time"
)

type tokenBucket struct {
	MaxBandWidth int // In kbps
	tokens       int // One token is 100 kb
	timerGranularity int
	mutex        *sync.Mutex
}

func (tb *tokenBucket) start() {

	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.tokens = tb.MaxBandWidth

	timer1 := time.NewTimer(10 * time.Millisecond)

	go func() {
		for {
			<-timer1.C
			// TODO for some reason I can't import math and can't use the max function. Maybe this resolves itself magically in the future
			tb.mutex.Lock()
			if(tb.tokens+(tb.MaxBandWidth/100) > tb.MaxBandWidth) {
				tb.tokens = tb.MaxBandWidth
			} else {
				tb.tokens = tb.tokens+(tb.MaxBandWidth/100)
			}
			tb.mutex.Unlock()
		}
    }()
}

func (pq *packetQueue) police(qp *qPkt) policeAction {
	pq.tb.mutex.Lock()
	defer pq.tb.mutex.Unlock()

	packetSize := (qp.rp.Bytes().Len() / 1024)

	if(pq.tb.tokens - packetSize > 0) {
		pq.tb.tokens = pq.tb.tokens - packetSize
	} else {
		qp.act.action = DROP
		qp.act.reason = BandWidthExceeded
	}
	return qp.act.action
}

func (qp *qPkt) sendNotification() {
	select {
	case r.notifications <- qp:
	default:
	}
}
