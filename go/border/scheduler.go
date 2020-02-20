package main

import "github.com/scionproto/scion/go/lib/log"

func (r *Router) dequeue(i int) {

	length := r.queues[i].getLength()
	log.Debug("The queue has length", "length", length)

	if length > 0 {
		qps := r.queues[i].popMultiple(length)
		for _, qp := range qps {
			r.forwardPacket(qp.rp)
		}
	}
}

func (r *Router) dequeuer() {
	for {
		j := <-r.flag
		i := 0

		for i < len(r.queues) {
			r.dequeue((j + i) % (len(r.queues)))
			i = i + 1
		}
	}
}

func (r *Router) wrrDequer() {

	i := 0
	qsum := 0
	for i < len(r.queues) {
		qsum = qsum + r.queues[i].priority
	}

	for {
		j := <-r.flag
		i := 0

		for i < len(r.queues) {
			r.wrrDequeue((j+i)%(len(r.queues)), 1)
			i = i + 1
		}
	}
}

func (r *Router) wrrDequeue(queueNo int, qsum int) {

	length := r.queues[queueNo].getLength()
	pktToDequeue := min(64*(r.queues[queueNo].priority/qsum), 1)

	log.Debug("The queue has length", "length", length)
	log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	if length > 0 {
		qps := r.queues[queueNo].popMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			r.forwardPacket(qp.rp)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
