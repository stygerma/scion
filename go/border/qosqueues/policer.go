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
	MaxBandWidth int // In bps
	tokens       int // One token is 1 b
	tokenSpent   int
	lastRefill   time.Time
	mutex        *sync.Mutex
	CurrBW       uint64
}

func (tb *tokenBucket) refill() {

	now := time.Now()

	timeSinceLastUpdate := now.Sub(tb.lastRefill).Milliseconds()

	log.Trace("Last update was", "ms ago", timeSinceLastUpdate)

	if timeSinceLastUpdate > 100 {

		newTokens := ((tb.MaxBandWidth) * int(timeSinceLastUpdate)) / (1000)
		tb.lastRefill = now

		log.Trace("Add new tokens", "#tokens", newTokens)
		log.Trace("On Update: Spent token in last period", "#tokens", tb.tokenSpent)

		tb.CurrBW = uint64(tb.tokenSpent/int(timeSinceLastUpdate)) * 1000

		tb.tokenSpent = 0

		if tb.tokens+newTokens > tb.MaxBandWidth {
			tb.tokens = tb.MaxBandWidth
		} else {
			tb.tokens += newTokens
		}
	}

}
