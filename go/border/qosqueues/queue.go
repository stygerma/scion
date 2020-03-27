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

	"github.com/scionproto/scion/go/border/rpkt"
)

type PoliceAction int

const (
	// PASS Pass the packet
	PASS PoliceAction = iota
	// NOTIFY Notify the sending host of the packet
	NOTIFY
	// DROP Drop the packet
	DROP
	// DROPNOTIFY Drop and then notify someone
	DROPNOTIFY
)

type Violation int

const (
	None Violation = iota
	BandWidthExceeded
	queueFull
)

// Action is
type Action struct {
	reason Violation
	action PoliceAction
}

type QPkt struct {
	QueueNo int
	Act     Action
	Rp      *rpkt.RtrPkt
}

type ActionProfile struct {
	FillLevel int
	Prob      int
	Action    PoliceAction
}

type PacketQueue struct {
	Name         string
	ID           int
	MinBandwidth int
	MaxBandWidth int
	PoliceRate   int
	MaxLength    int
	Priority     int
	Profile      []ActionProfile
}

type PacketQueueInterface interface {
	InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex)
	Enqueue(*QPkt)
	Pop() *QPkt
	PopMultiple(number int) []*QPkt
	GetFillLevel() int
	GetLength() int
	CheckAction() PoliceAction
	Police(qp *QPkt, shouldLog bool) PoliceAction
	GetPriority() int
	GetMinBandwidth() int
	GetPacketQueue() PacketQueue
}

func ReturnAction(polAction PoliceAction, profAction PoliceAction) PoliceAction {

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