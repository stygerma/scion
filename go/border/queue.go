package main

import (
	"sync"

	"github.com/scionproto/scion/go/border/rpkt"
)

// Queue is a single queue
type packetQueue struct {
	// Id string

	mutex *sync.Mutex

	queue     []*rpkt.RtrPkt
	length int
	maxLength int
	priority  int
}

func (pq *packetQueue) enqueue(rp *rpkt.RtrPkt) {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queue = append(pq.queue, rp)
	pq.length = pq.length + 1

}

func (pq *packetQueue) canDequeue() bool {

	return pq.length > -1
}

func (pq *packetQueue) getLength() int {

	return pq.length
}

func (pq *packetQueue) peek() *rpkt.RtrPkt {

	return pq.queue[0]
}

func (pq *packetQueue) pop() *rpkt.RtrPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[0]
	pq.queue = pq.queue[1:]
	pq.length = pq.length - 1

	return pkt
}