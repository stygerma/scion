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

	"github.com/scionproto/scion/go/border/qos/qosconf"
	"github.com/scionproto/scion/go/border/rpkt"
)

type QPkt struct {
	QueueNo int
	Act     Action
	Rp      *rpkt.RtrPkt
}

type NPkt struct {
	Rule *InternalClassRule
	Qpkt *QPkt
}

type Violation int

const (
	None Violation = iota
	BandWidthExceeded
	queueFull
)

// Action is
type Action struct {
	rule   *InternalClassRule
	reason Violation
	action qosconf.PoliceAction
}

type ActionProfile struct {
	FillLevel int                  `yaml:"fill-level"`
	Prob      int                  `yaml:"prob"`
	Action    qosconf.PoliceAction `yaml:"action"`
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
	CheckAction() qosconf.PoliceAction
	Police(qp *QPkt) qosconf.PoliceAction
	GetPriority() int
	GetMinBandwidth() int
	GetMaxBandwidth() int
	GetPacketQueue() PacketQueue
}

func ReturnActionOld(polAction qosconf.PoliceAction, profAction qosconf.PoliceAction) qosconf.PoliceAction {

	if polAction == qosconf.DROPNOTIFY || profAction == qosconf.DROPNOTIFY {
		return qosconf.DROPNOTIFY
	}

	if polAction == qosconf.DROP || profAction == qosconf.DROP {
		return qosconf.DROP
	}

	if polAction == qosconf.NOTIFY || profAction == qosconf.NOTIFY {
		return qosconf.NOTIFY
	}

	return qosconf.PASS
}

func ReturnAction(polAction qosconf.PoliceAction, profAction qosconf.PoliceAction) qosconf.PoliceAction {

	pol, prof := 3-polAction, 3-profAction
	if pol*prof == 0 {
		return qosconf.DROPNOTIFY
	}
	pol--
	prof--
	if pol*prof == 0 {
		return qosconf.DROP
	}
	pol--
	prof--
	if pol*prof == 0 {
		return qosconf.NOTIFY
	}
	return qosconf.PASS
}
