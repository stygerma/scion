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

type packetQueue struct {
	Name         string          `yaml:"name"`
	ID           int             `yaml:"id"`
	MinBandwidth int             `yaml:"CIR"`
	MaxBandWidth int             `yaml:"PIR"`
	PoliceRate   int             `yaml:"policeRate"`
	MaxLength    int             `yaml:"maxLength"`
	priority     int             `yaml:"priority"`
	Profile      []actionProfile `yaml:"profile"`

	mutex *sync.Mutex

	queue  []*qPkt
	length int
	tb     tokenBucket
}

// TODO: Implement this. It currently does nothing
type actionProfile struct {
	FillLevel int          `yaml:"fill-level"`
	Prob      int          `yaml:"prob"`
	Action    policeAction `yaml:"action"`
}

func (pq *packetQueue) enqueue(rp *qPkt) {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queue = append(pq.queue, rp)
	pq.length = pq.length + 1

}

func (pq *packetQueue) canDequeue() bool {

	return pq.length > 0
}

func (pq *packetQueue) getLength() int {

	return pq.length
}

func (pq *packetQueue) peek() *qPkt {

	return pq.queue[0]
}

func (pq *packetQueue) pop() *qPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[0]
	pq.queue = pq.queue[1:]
	pq.length = pq.length - 1

	return pkt
}

func (pq *packetQueue) popMultiple(number int) []*qPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[:number]
	pq.queue = pq.queue[number:]
	pq.length = pq.length - number

	return pkt
}
