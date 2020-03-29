// Copyright 2020 ETH Zurich
// Copyright 2020 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qosqueues

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/lib/log"
)

type tokenBucket struct {
	maxBandWidth int // In Bps
	tokens       int // One token is 1 B
	tokenSpent   int
	lastRefill   time.Time
	mutex        *sync.Mutex
	CurrBW       uint64
}

func (tb *tokenBucket) Init(maxBandwidth int) {
	tb.maxBandWidth = maxBandwidth
	tb.tokens = maxBandwidth
	tb.tokenSpent = 0
	tb.lastRefill = time.Now()
	tb.mutex = &sync.Mutex{}
}

// Only call this if you have a lock on tb!
func (tb *tokenBucket) refill() {

	// tb.mutex.Lock()
	// defer tb.mutex.Unlock()

	log.Trace("Overall available bandwidth per second", "MaxBandWidth", tb.maxBandWidth)
	log.Trace("Spent token in last period", "#tokens", tb.tokenSpent)
	log.Trace("Available bandwidth before refill", "bandwidth", tb.tokens)

	now := time.Now()

	timeSinceLastUpdate := now.Sub(tb.lastRefill).Milliseconds()

	log.Trace("Last update was", "ms ago", timeSinceLastUpdate)

	if timeSinceLastUpdate > 100 {

		newTokens := ((tb.maxBandWidth) * int(timeSinceLastUpdate)) / (1000)
		tb.lastRefill = now

		log.Trace("Add new tokens", "#tokens", newTokens)
		log.Trace("On Update: Spent token in last period", "#tokens", tb.tokenSpent)

		tb.CurrBW = uint64(tb.tokenSpent/int(timeSinceLastUpdate)) * 1000

		tb.tokenSpent = 0

		if tb.tokens+newTokens > tb.maxBandWidth {
			tb.tokens = tb.maxBandWidth
		} else {
			tb.tokens += newTokens
		}
	}

	log.Trace("Available bandwidth after refill", "bandwidth", tb.tokens)

}

func (tb *tokenBucket) PoliceBucket(qp *QPkt) PoliceAction {

	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	packetSize := (qp.Rp.Bytes().Len()) // In byte

	tokenForPacket := packetSize // In byte

	tb.refill()
	log.Trace("Tokens necessary for packet", "tokens", tokenForPacket)
	log.Trace("Tokens necessary for packet", "bytes", qp.Rp.Bytes().Len())

	if tb.tokens-tokenForPacket > 0 {
		tb.tokens = tb.tokens - tokenForPacket
		tb.tokenSpent += tokenForPacket
		qp.Act.action = PASS
		qp.Act.reason = None
	} else {
		qp.Act.action = DROP
		qp.Act.reason = BandWidthExceeded
	}

	log.Trace("Available bandwidth after update", "bandwidth", tb.tokens)

	return qp.Act.action

}
