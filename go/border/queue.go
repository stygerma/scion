package main

import (
	"sync"

	"github.com/scionproto/scion/go/border/rpkt"
)

type qPkt struct {
	queueNo int
	rp *rpkt.RtrPkt
}

// Queue is a single queue
type packetQueue struct {
	// Id string

	mutex *sync.Mutex

	queue     []*qPkt
	length int
	maxLength int
	priority  int
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