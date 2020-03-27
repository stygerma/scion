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

import (
	"sync"
	"testing"
	"time"

	"github.com/scionproto/scion/go/border/rpkt"
)

var packets = make(chan *rpkt.RtrPkt, 100)

/*
Things to do:

1. Set up router with a topology
2. Create a packet

*/

func setupQueue() packetQueue {

	bandwidth := 0
	priority := 1

	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	que := packetQueue{MaxLength: 128, MinBandwidth: priority, MaxBandWidth: priority, mutex: &sync.Mutex{}, tb: bucket}

	return que

}

func setupQueuePaket() qPkt {

	return qPkt{queueNo: 0, rp: nil}
}

func TestBasicEnqueue(t *testing.T) {
	que := setupQueue()
	pkt := setupQueuePaket()
	que.enqueue(&pkt)
	length := que.getLength()
	if length != 1 {
		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
	}
}

func TestMultiEnqueue(t *testing.T) {
	que := setupQueue()
	pkt := setupQueuePaket()
	j := 15
	i := 0

	for i < j {
		que.enqueue(&pkt)
		i++
	}
	length := que.getLength()

	if length != j {
		t.Errorf("Enqueue one packet should give length %d gave length %d", j, length)
	}
}

func BenchmarkEnqueue(b *testing.B) {
	que := setupQueue()
	pkt := setupQueuePaket()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		que.enqueue(&pkt)
	}
}

func benchmarkPop(popNo int, b *testing.B) {
	que := setupQueue()
	pkt := setupQueuePaket()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		j := 0
		for j < popNo {
			que.enqueue(&pkt)
			j++
		}
		b.StartTimer()
		que.popMultiple(popNo)
	}
}

func BenchmarkPop1(b *testing.B) { benchmarkPop(1, b) }
func BenchmarkPop5(b *testing.B) { benchmarkPop(10, b) }

func TestCallPacketGen(t *testing.T) {
	_ = rpkt.JFPrepareRtrPacketSample(t)
}

func TestBasicRoute(t *testing.T) {

	r, _ := setupTestRouter(t)
	r.initQueueing("sample-config.yaml")

	r.forwarder = r.forwardPacketTest

	rp := rpkt.JFPrepareRtrPacketSample(t)

	r.queuePacket(rp)

	time.Sleep(2 * time.Second)

	select {
	case returnedPacket := <-packets:
		if returnedPacket != rp {
			t.Errorf("Returned wrong packet!")
		} else {
			t.Log("We got the packet back 🥳👯‍♂️👯‍♀️")
		}
	default:
		t.Errorf("Returned no packet!")
	}

}

func (r *Router) forwardPacketTest(rp *rpkt.RtrPkt) {

	// defer rp.Release()

	packets <- rp

}

func TestHundredPacketQueue(t *testing.T) {

	r, _ := setupTestRouter(t)

	r.initQueueing("sample-config.yaml")
	r.forwarder = r.forwardPacketTest

	ps := make([]*rpkt.RtrPkt, 100)

	for i := 0; i < 100; i++ {
		rp := rpkt.JFPrepareRtrPacketSample(t)
		r.queuePacket(rp)
		ps[i] = rp
	}

	time.Sleep(2 * time.Second)

	for i := 0; i < 100; i++ {
		select {
		case returnedPacket := <-packets:
			if returnedPacket != ps[i] {
				t.Errorf("Returned wrong packet!")
			} else {
				t.Log("We got the packet back 🥳👯‍♂️👯‍♀️")
			}
		default:
			t.Errorf("Returned no packet!")
		}
	}
}
