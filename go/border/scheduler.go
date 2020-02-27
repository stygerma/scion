package main

import "github.com/scionproto/scion/go/lib/log"

func (r *Router) dequeue(i int) {

	length := r.queues[i].getLength()
	log.Debug("The queue has length", "length", length)

	if length > 0 {
		qps := r.queues[i].popMultiple(length)
		for _, qp := range qps {
			r.forwarder(qp.rp)
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

func (r *Router) drrDequer() {

	i := 0
	qsum := 0
	for i < len(r.queues) {
		qsum = qsum + r.queues[i].priority
		i++
	}

	for {
		j := <-r.flag
		i := 0

		for i < len(r.queues) {
			r.drrMinMaxDequeue((j+i)%(len(r.queues)), 1)
			i++
		}
	}
}

func (r *Router) drrDequeue(queueNo int, qsum int) {

	length := r.queues[queueNo].getLength()
	pktToDequeue := min(64*(r.queues[queueNo].priority/qsum), 1)

	log.Debug("The queue has length", "length", length)
	log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	if length > 0 {
		qps := r.queues[queueNo].popMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			r.forwarder(qp.rp)
		}
	}
}

func (r *Router) drrMinMaxDequeue(queueNo int, qsum int) {

	length := r.queues[queueNo].getLength()
	pktToDequeue := min(64*(r.queues[queueNo].minBandwidth/qsum), 1)

	log.Debug("The queue has length", "length", length)
	log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	if length > 0 {

		if r.surplusAvailable() {
			log.Debug("Surplus available", "surplus", r.schedulerSurplus)
			if length > pktToDequeue {
				pktToDequeue = r.getFromSurplus(queueNo, length)
				log.Debug("Dequeueing above minimum", "quantum", pktToDequeue)
			} else {
				if pktToDequeue-length > 0 {
					r.payIntoSurplus(queueNo, pktToDequeue-length)
					log.Debug("Paying into surplus", "payment", pktToDequeue-length)
				}
			}
		}

		qps := r.queues[queueNo].popMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			r.forwarder(qp.rp)
		}
	}
}

func (r *Router) getFromSurplus(queueNo int, request int) int {

	r.schedulerSurplusMtx.Lock()
	defer r.schedulerSurplusMtx.Unlock()

	// Check limit for queue
	// Take out of surplus

	i := 0
	qsum := 0
	for i < len(r.queues) {
		qsum = qsum + r.queues[i].minBandwidth
		i++
	}
	upperLimit := min(64*(r.queues[queueNo].maxBandwidth/qsum), 1)

	credit := min(r.schedulerSurplus, upperLimit)

	r.schedulerSurplus = r.schedulerSurplus - credit

	return credit

}

func (r *Router) payIntoSurplus(queueNo int, payment int) {

	r.schedulerSurplusMtx.Lock()
	defer r.schedulerSurplusMtx.Unlock()

	r.schedulerSurplus = r.schedulerSurplus + payment

}

func (r *Router) surplusAvailable() bool {

	r.schedulerSurplusMtx.Lock()
	defer r.schedulerSurplusMtx.Unlock()

	return r.schedulerSurplus > 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
