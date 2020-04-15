// Copyright 2020 ETH Zurich
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

package scheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
)

type SchedulerInterface interface {
	Init(routerConfig queues.InternalRouterConfig)
	Dequeuer(routerConfig queues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt))
	Dequeue(queue queues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int)
	GetMessages() *chan bool
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var iterations int
var t0 time.Time
var tokensUsed [5]int
var cirTokens [5]int
var pirTokens [5]int
var payedIntoSurplus [5]int
var forceTake [5]int
var overallTokensUsed int
