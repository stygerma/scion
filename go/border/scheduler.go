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

import "github.com/scionproto/scion/go/lib/log"

// This is a standard round robin dequeue ignoring things like priority

func (r *Router) dequeue(i int) {

	length := r.config.Queues[i].getLength()
	log.Debug("The queue has length", "length", length)

	if length > 0 {
		qps := r.config.Queues[i].popMultiple(length)
		for _, qp := range qps {
			r.forwarder(qp.rp)
		}
	}
}

func (r *Router) dequeuer() {
	for {
		j := <-r.flag
		i := 0

		for i < len(r.config.Queues) {
			r.dequeue((j + i) % (len(r.config.Queues)))
			i = i + 1
		}
	}
}

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

func (r *Router) drrDequer() {

	i := 0
	qsum := 0
	for i < len(r.config.Queues) {
		qsum = qsum + r.config.Queues[i].getPriority()
		i++
	}

	for {
		j := <-r.flag
		i := 0

		for i < len(r.config.Queues) {
			r.drrDequeue((j+i)%(len(r.config.Queues)), 1)
			i++
		}
	}
}

func (r *Router) drrDequeue(queueNo int, qsum int) {

	length := r.config.Queues[queueNo].getLength()
	pktToDequeue := min(64*(r.config.Queues[queueNo].getPriority()/qsum), 1)

	log.Debug("The queue has length", "length", length)
	log.Debug("Dequeueing packets", "quantum", pktToDequeue)

	if length > 0 {
		qps := r.config.Queues[queueNo].popMultiple(max(length, pktToDequeue))
		for _, qp := range qps {
			r.forwarder(qp.rp)
		}
	}
}

// This is also a deficit round robin dequeuer. But instead of the priority field it uses the min-bandwidth field for the minimum number of packets to dequeue. If there are fewer than the minimal value of packets to dequeue, the remaining min-bandwidth will be put onto a surplus counter and another queue might use more than its min-bandwidth (but still less than its max-bandwidth).

func (r *Router) drrMinMaxDequer() {

	i := 0
	qsum := 0
	for i < len(r.config.Queues) {
		qsum = qsum + r.config.Queues[i].getPriority()
		i++
	}

	for {
		j := <-r.flag
		i := 0

		for i < len(r.config.Queues) {
			r.drrMinMaxDequeue((j+i)%(len(r.config.Queues)), 1)
			i++
		}
	}
}

func (r *Router) drrMinMaxDequeue(queueNo int, qsum int) {

	length := r.config.Queues[queueNo].getLength()
	pktToDequeue := min(64*(r.config.Queues[queueNo].getMinBandwidth()/qsum), 1)

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

		qps := r.config.Queues[queueNo].popMultiple(max(length, pktToDequeue))
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
	for i < len(r.config.Queues) {
		qsum = qsum + r.config.Queues[i].getMinBandwidth()
		i++
	}
	upperLimit := min(64*(r.config.Queues[queueNo].getMinBandwidth()/qsum), 1)

	credit := min(r.schedulerSurplus.surplus, upperLimit)

	r.schedulerSurplus.surplus = r.schedulerSurplus.surplus - credit

	return credit

}

func (r *Router) payIntoSurplus(queueNo int, payment int) {

	r.schedulerSurplusMtx.Lock()
	defer r.schedulerSurplusMtx.Unlock()

	r.schedulerSurplus.surplus = min(r.schedulerSurplus.surplus+(payment-r.schedulerSurplus.payments[queueNo]), 0)
	r.schedulerSurplus.payments[queueNo] = payment

}

func (r *Router) surplusAvailable() bool {

	r.schedulerSurplusMtx.Lock()
	defer r.schedulerSurplusMtx.Unlock()

	return r.schedulerSurplus.surplus > 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
