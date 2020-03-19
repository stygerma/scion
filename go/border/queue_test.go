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

func setupQueue() packetSliceQueue {

	bandwidth := 0
	priority := 1

	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	que := packetSliceQueue{MaxLength: 128, MinBandwidth: priority, MaxBandWidth: priority, mutex: &sync.Mutex{}, tb: bucket}

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

func Benchmark600Enqueue(b *testing.B) {
	que := setupQueue()
	pkt := setupQueuePaket()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 600; j++ {
			que.enqueue(&pkt)
		}
		pkts := que.popMultiple(100)
		_ = pkts
	}
}

func Benchmark600BufEnqueue(b *testing.B) {

	bandwidth := 0
	priority := 1

	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	que := packetBufQueue{MaxLength: 600, MinBandwidth: priority, MaxBandWidth: priority, mutex: &sync.Mutex{}, tb: bucket}
	que.initQueue(&sync.Mutex{}, &sync.Mutex{})

	pkt := setupQueuePaket()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 600; j++ {
			que.enqueue(&pkt)
		}
		pkts := que.popMultiple(100)
		_ = pkts
	}
}

func BenchmarkEnqDeque(b *testing.B) {

	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	pkt := setupQueuePaket()

	benchmarks := []struct {
		name       string
		noPackets  int
		dequeueDiv int
		que        packetQueue
	}{
		{"Buf Queue 80 Packets", 80, 8, &packetBufQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 80 Packets", 80, 8, &packetSliceQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 600 Packets", 600, 6, &packetBufQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 600 Packets", 600, 6, &packetSliceQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 7000 Packets", 7000, 7, &packetBufQueue{MaxLength: 7000, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 7000 Packets", 7000, 7, &packetSliceQueue{MaxLength: 7000, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.que.initQueue(&sync.Mutex{}, &sync.Mutex{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for i := 0; i < bm.noPackets; i++ {
					bm.que.enqueue(&pkt)
				}
				for i := 0; i < bm.dequeueDiv; i++ {
					bm.que.popMultiple(bm.noPackets / bm.dequeueDiv)
				}
			}
		})
	}
}

func BenchmarkEnqDequeMult(b *testing.B) {

	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	pkt := setupQueuePaket()

	benchmarks := []struct {
		name                  string
		noPacketsPerIteration int
		iterations            int
		que                   packetQueue
	}{
		{"Slice Queue 10 Packets 100 times", 10, 10, &packetSliceQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 10 Packets 100 times", 10, 10, &packetBufQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Custom Queue 10 Packets 100 times", 10, 10, &customPacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 80 Packets 8 times", 80, 8, &packetSliceQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 80 Packets 8 times", 80, 8, &packetBufQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Custom Queue 80 Packets 8 times", 80, 8, &customPacketQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 600 Packets 6 times", 600, 6, &packetSliceQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 600 Packets 6 times", 600, 6, &packetBufQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Custom Queue 600 Packets 6 times", 600, 6, &customPacketQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.que.initQueue(&sync.Mutex{}, &sync.Mutex{})
			var pkts []*qPkt
			_ = pkts
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for i := 0; i < bm.iterations; i++ {
					for i := 0; i < bm.noPacketsPerIteration; i++ {
						bm.que.enqueue(&pkt)
					}
					pkts = bm.que.popMultiple(bm.noPacketsPerIteration)
				}

			}
		})
	}
}

func BenchmarkEnqOnly(b *testing.B) {

	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	pkt := setupQueuePaket()

	benchmarks := []struct {
		name                  string
		noPacketsPerIteration int
		iterations            int
		que                   packetQueue
	}{
		{"Slice Queue 10 Packets 100 times", 10, 10, &packetSliceQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 10 Packets 100 times", 10, 10, &packetBufQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Custom Queue 10 Packets 100 times", 10, 10, &customPacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Channel Queue 10 Packets 100 times", 10, 10, &channelPacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}}}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.que.initQueue(&sync.Mutex{}, &sync.Mutex{})
			var pkts []*qPkt
			_ = pkts
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.que.enqueue(&pkt)
				_ = bm.que.pop()
			}
		})
	}
}

func BenchmarkEnqPop(b *testing.B) {

	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	pkt := setupQueuePaket()

	benchmarks := []struct {
		name                  string
		noPacketsPerIteration int
		iterations            int
		que                   packetQueue
	}{
		{"Slice Queue 10 Packets 100 times", 10, 10, &packetSliceQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 10 Packets 100 times", 10, 10, &packetBufQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Custom Queue 10 Packets 100 times", 10, 10, &customPacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.que.initQueue(&sync.Mutex{}, &sync.Mutex{})
			var pkts []*qPkt
			_ = pkts
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.que.enqueue(&pkt)
				_ = bm.que.pop()
			}
		})
	}
}

func BenchmarkEnqPopMult(b *testing.B) {

	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	pkt := setupQueuePaket()

	benchmarks := []struct {
		name                  string
		noPacketsPerIteration int
		iterations            int
		que                   packetQueue
	}{
		{"Slice Queue 10 Packets 100 times", 10, 10, &packetSliceQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 10 Packets 100 times", 10, 10, &packetBufQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Custom Queue 10 Packets 100 times", 10, 10, &customPacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.que.initQueue(&sync.Mutex{}, &sync.Mutex{})
			var pkts []*qPkt
			_ = pkts
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.que.enqueue(&pkt)
				bm.que.enqueue(&pkt)
				bm.que.enqueue(&pkt)
				_ = bm.que.popMultiple(3)
			}
		})
	}
}

func BenchmarkEnqDequeSingleMult(b *testing.B) {

	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	pkt := setupQueuePaket()

	benchmarks := []struct {
		name                  string
		noPacketsPerIteration int
		iterations            int
		que                   packetQueue
	}{
		{"Buf Queue 64 Packets", 64, 8, &packetBufQueue{MaxLength: 64, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 64 Packets", 64, 8, &packetSliceQueue{MaxLength: 64, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 512 Packets", 512, 6, &packetBufQueue{MaxLength: 512, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 512 Packets", 512, 6, &packetSliceQueue{MaxLength: 512, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Buf Queue 8192 Packets", 8192, 7, &packetBufQueue{MaxLength: 8192, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
		{"Slice Queue 8192 Packets", 8192, 7, &packetSliceQueue{MaxLength: 8192, MinBandwidth: 0, MaxBandWidth: 0, mutex: &sync.Mutex{}, tb: bucket}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.que.initQueue(&sync.Mutex{}, &sync.Mutex{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for i := 0; i < bm.iterations; i++ {
					for i := 0; i < bm.noPacketsPerIteration; i++ {
						bm.que.enqueue(&pkt)
					}
					for i := 0; i < bm.noPacketsPerIteration; i++ {
						pkts := bm.que.pop()
						_ = pkts
					}

				}

			}
		})
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

func TestHundredpacketSliceQueue(t *testing.T) {

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
