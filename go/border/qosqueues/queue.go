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
	PASS PoliceAction = 0
	// NOTIFY Notify the sending host of the packet
	NOTIFY PoliceAction = 1
	// DROP Drop the packet
	DROP PoliceAction = 2
	// DROPNOTIFY Drop and then notify someone
	DROPNOTIFY PoliceAction = 3
)

type Violation int

const (
	// None none
	None = 0
	// BandWidthExceeded ...
	BandWidthExceeded = 1
	// queueFull
	queueFull = 2
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

type actionProfile struct {
	FillLevel int          `yaml:"fill-level"`
	Prob      int          `yaml:"prob"`
	Action    PoliceAction `yaml:"action"`
}

type PacketQueue struct {
	Name         string          `yaml:"name"`
	ID           int             `yaml:"id"`
	MinBandwidth int             `yaml:"CIR"`
	MaxBandWidth int             `yaml:"PIR"`
	PoliceRate   int             `yaml:"policeRate"`
	MaxLength    int             `yaml:"maxLength"`
	priority     int             `yaml:"priority"`
	Profile      []actionProfile `yaml:"profile"`
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
