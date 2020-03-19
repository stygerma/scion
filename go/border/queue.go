// Copyright 2020 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
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

package main

import (
	"sync"

	"github.com/scionproto/scion/go/border/rpkt"
)

type policeAction int

const (
	// PASS Pass the packet
	PASS policeAction = 0
	// NOTIFY Notify the sending host of the packet
	NOTIFY policeAction = 1
	// DROP Drop the packet
	DROP policeAction = 2
	// DROPNOTIFY Drop and then notify someone
	DROPNOTIFY policeAction = 3
)

type violation int

const (
	// None none
	None = 0
	// BandWidthExceeded ...
	BandWidthExceeded = 1
	// queueFull
	queueFull = 2
)

type action struct {
	reason violation
	action policeAction
}

type qPkt struct {
	queueNo int
	act     action
	rp      *rpkt.RtrPkt
}

type actionProfile struct {
	FillLevel int          `yaml:"fill-level"`
	Prob      int          `yaml:"prob"`
	Action    policeAction `yaml:"action"`
}

type packetQueue interface {
	initQueue(mutQue *sync.Mutex, mutTb *sync.Mutex)
	enqueue(*qPkt)
	pop() *qPkt
	popMultiple(number int) []*qPkt
	getFillLevel() int
	getLength() int
	checkAction() policeAction
	police(qp *qPkt, shouldLog bool) policeAction
	getPriority() int
	getMinBandwidth() int
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
