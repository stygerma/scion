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
