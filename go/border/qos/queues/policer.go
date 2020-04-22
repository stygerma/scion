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

package queues

import (
	"sync"
	"time"

	"github.com/scionproto/scion/go/border/qos/conf"
)

type tokenBucket struct {
	maxBandWidth int // In Bps
	tokens       int // One token is 1 B
	lastRefill   time.Time
	mutex        *sync.Mutex
}

func (tb *tokenBucket) Init(maxBandwidth int) {
	tb.maxBandWidth = maxBandwidth
	tb.tokens = maxBandwidth
	tb.lastRefill = time.Now()
	tb.mutex = &sync.Mutex{}
}

func (tb *tokenBucket) refill() {

	now := time.Now()

	timeSinceLastUpdate := now.Sub(tb.lastRefill).Milliseconds()

	if timeSinceLastUpdate > 100 {

		newTokens := ((tb.maxBandWidth) * int(timeSinceLastUpdate)) / (1000)
		tb.lastRefill = now

		if tb.tokens+newTokens > tb.maxBandWidth {
			tb.tokens = tb.maxBandWidth
		} else {
			tb.tokens += newTokens
		}
	}
}

func (tb *tokenBucket) PoliceBucket(qp *QPkt) conf.PoliceAction {

	tokenForPacket := (qp.Rp.Bytes().Len()) // In byte

	tb.refill()

	if tb.tokens-tokenForPacket > 0 {
		tb.tokens = tb.tokens - tokenForPacket
		qp.Act.action = conf.PASS
		qp.Act.reason = None
	} else {
		qp.Act.action = conf.DROP
		qp.Act.reason = BandWidthExceeded
	}

	return qp.Act.action

}
