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
	"github.com/scionproto/scion/go/border/rpkt"
)

var packets = make(chan *rpkt.RtrPkt, 100)

/*
Things to do:

1. Set up router with a topology
2. Create a packet

*/

// func setupQueue() PacketSliceQueue {

// 	bandwidth := 0
// 	priority := 1

// 	pq := PacketQueue{MaxLength: 128, MinBandwidth: priority, MaxBandWidth: priority}

// 	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	que := PacketSliceQueue{pktQue: pq, mutex: &sync.Mutex{}, tb: bucket}

// 	return que

// }

// func setupQueuePaket() QPkt {

// 	return QPkt{QueueNo: 0, Rp: nil}
// }

// func TestBasicEnqueue(t *testing.T) {
// 	que := setupQueue()
// 	pkt := setupQueuePaket()
// 	que.Enqueue(&pkt)
// 	length := que.GetLength()
// 	if length != 1 {
// 		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
// 	}
// }

// func TestMultiEnqueue(t *testing.T) {
// 	que := setupQueue()
// 	pkt := setupQueuePaket()
// 	j := 15
// 	i := 0

// 	for i < j {
// 		que.Enqueue(&pkt)
// 		i++
// 	}
// 	length := que.GetLength()

// 	if length != j {
// 		t.Errorf("Enqueue one packet should give length %d gave length %d", j, length)
// 	}
// }

// func BenchmarkEnqueue(b *testing.B) {
// 	que := setupQueue()
// 	pkt := setupQueuePaket()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		que.Enqueue(&pkt)
// 	}
// }

// func Benchmark600Enqueue(b *testing.B) {
// 	que := setupQueue()
// 	pkt := setupQueuePaket()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		for j := 0; j < 600; j++ {
// 			que.Enqueue(&pkt)
// 		}
// 		pkts := que.PopMultiple(100)
// 		_ = pkts
// 	}
// }

// func Benchmark600BufEnqueue(b *testing.B) {

// 	bandwidth := 0
// 	priority := 1

// 	pq := PacketQueue{MaxLength: 600, MinBandwidth: priority, MaxBandWidth: priority}
// 	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	que := PacketBufQueue{pktQue: pq, mutex: &sync.Mutex{}, tb: bucket}
// 	que.InitQueue(pq, &sync.Mutex{}, &sync.Mutex{})

// 	pkt := setupQueuePaket()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		for j := 0; j < 600; j++ {
// 			que.Enqueue(&pkt)
// 		}
// 		pkts := que.PopMultiple(100)
// 		_ = pkts
// 	}
// }

// func BenchmarkEnqDeque(b *testing.B) {

// 	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	pkt := setupQueuePaket()

// 	pq1 := PacketQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0}
// 	pq2 := PacketQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0}
// 	pq3 := PacketQueue{MaxLength: 7000, MinBandwidth: 0, MaxBandWidth: 0}

// 	benchmarks := []struct {
// 		name          string
// 		noPackets     int
// 		dequeueDiv    int
// 		que           PacketQueueInterface
// 		internalQueue PacketQueue
// 	}{
// 		{"Ringbuf Queue 80 Packets", 80, 8, &PacketBufQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Slice Queue 80 Packets", 80, 8, &PacketSliceQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Ringbuf Queue 600 Packets", 600, 6, &PacketBufQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Slice Queue 600 Packets", 600, 6, &PacketSliceQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Ringbuf Queue 7000 Packets", 7000, 7, &PacketBufQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Slice Queue 7000 Packets", 7000, 7, &PacketSliceQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 	}

// 	for _, bm := range benchmarks {
// 		b.Run(bm.name, func(b *testing.B) {
// 			bm.que.InitQueue(bm.internalQueue, &sync.Mutex{}, &sync.Mutex{})
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				for i := 0; i < bm.noPackets; i++ {
// 					bm.que.Enqueue(&pkt)
// 				}
// 				for i := 0; i < bm.dequeueDiv; i++ {
// 					bm.que.PopMultiple(bm.noPackets / bm.dequeueDiv)
// 				}
// 			}
// 		})
// 	}
// }

// func BenchmarkEnqDequeMult(b *testing.B) {

// 	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	pkt := setupQueuePaket()

// 	pq1 := PacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0}
// 	pq2 := PacketQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0}
// 	pq3 := PacketQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0}

// 	benchmarks := []struct {
// 		name                  string
// 		noPacketsPerIteration int
// 		iterations            int
// 		que                   PacketQueueInterface
// 		internalPacketQueue   PacketQueue
// 	}{
// 		{"Slice Queue 10 Packets 100 times", 10, 10, &PacketSliceQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Ringbuf Queue 10 Packets 100 times", 10, 10, &PacketBufQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Custom Queue 10 Packets 100 times", 10, 10, &CustomPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Channel Queue 10 Packets 100 times", 10, 10, &ChannelPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Slice Queue 80 Packets 8 times", 80, 8, &PacketSliceQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Ringbuf Queue 80 Packets 8 times", 80, 8, &PacketBufQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Custom Queue 80 Packets 8 times", 80, 8, &CustomPacketQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Channel Queue 80 Packets 8 times", 80, 8, &ChannelPacketQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Slice Queue 600 Packets 6 times", 600, 6, &PacketSliceQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Ringbuf Queue 600 Packets 6 times", 600, 6, &PacketBufQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Custom Queue 600 Packets 6 times", 600, 6, &CustomPacketQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Channel Queue 600 Packets 6 times", 600, 6, &ChannelPacketQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 	}

// 	for _, bm := range benchmarks {
// 		b.Run(bm.name, func(b *testing.B) {
// 			bm.que.InitQueue(bm.internalPacketQueue, &sync.Mutex{}, &sync.Mutex{})
// 			var pkts []*QPkt
// 			_ = pkts
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				for i := 0; i < bm.iterations; i++ {
// 					for i := 0; i < bm.noPacketsPerIteration; i++ {
// 						bm.que.Enqueue(&pkt)
// 					}
// 					pkts = bm.que.PopMultiple(bm.noPacketsPerIteration)
// 				}

// 			}
// 		})
// 	}
// }

// func BenchmarkEnqDequeSingle(b *testing.B) {

// 	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	pkt := setupQueuePaket()

// 	pq1 := PacketQueue{MaxLength: 32, MinBandwidth: 0, MaxBandWidth: 0}
// 	pq2 := PacketQueue{MaxLength: 80, MinBandwidth: 0, MaxBandWidth: 0}
// 	pq3 := PacketQueue{MaxLength: 600, MinBandwidth: 0, MaxBandWidth: 0}

// 	benchmarks := []struct {
// 		name                  string
// 		noPacketsPerIteration int
// 		iterations            int
// 		que                   PacketQueueInterface
// 		internalPacketQueue   PacketQueue
// 	}{
// 		{"Slice Queue 10 Packets 1 times", 1, 1, &PacketSliceQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Ringbuf Queue 1 Packets 1 times", 1, 1, &PacketBufQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Custom Queue 1 Packets 1 times", 1, 1, &CustomPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Channel Queue 1 Packets 1 times", 1, 1, &ChannelPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Slice Queue 10 Packets 100 times", 10, 10, &PacketSliceQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Ringbuf Queue 10 Packets 100 times", 10, 10, &PacketBufQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Custom Queue 10 Packets 100 times", 10, 10, &CustomPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Channel Queue 10 Packets 100 times", 10, 10, &ChannelPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Slice Queue 80 Packets 8 times", 80, 8, &PacketSliceQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Ringbuf Queue 80 Packets 8 times", 80, 8, &PacketBufQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Custom Queue 80 Packets 8 times", 80, 8, &CustomPacketQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Channel Queue 80 Packets 8 times", 80, 8, &ChannelPacketQueue{pktQue: pq2, mutex: &sync.Mutex{}, tb: bucket}, pq2},
// 		{"Slice Queue 600 Packets 6 times", 600, 6, &PacketSliceQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Ringbuf Queue 600 Packets 6 times", 600, 6, &PacketBufQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Custom Queue 600 Packets 6 times", 600, 6, &CustomPacketQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 		{"Channel Queue 600 Packets 6 times", 600, 6, &ChannelPacketQueue{pktQue: pq3, mutex: &sync.Mutex{}, tb: bucket}, pq3},
// 	}

// 	for _, bm := range benchmarks {
// 		b.Run(bm.name, func(b *testing.B) {
// 			bm.que.InitQueue(bm.internalPacketQueue, &sync.Mutex{}, &sync.Mutex{})
// 			var pkts []*QPkt
// 			_ = pkts
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				for i := 0; i < bm.iterations; i++ {
// 					for i := 0; i < bm.noPacketsPerIteration; i++ {
// 						bm.que.Enqueue(&pkt)
// 					}
// 					for i := 0; i < bm.noPacketsPerIteration; i++ {
// 						_ = bm.que.Pop()
// 					}
// 				}

// 			}
// 		})
// 	}
// }

// func BenchmarkEnqPop(b *testing.B) {

// 	bucket := tokenBucket{MaxBandWidth: 0, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	pkt := setupQueuePaket()

// 	pq1 := PacketQueue{MaxLength: 10, MinBandwidth: 0, MaxBandWidth: 0}

// 	benchmarks := []struct {
// 		name                  string
// 		noPacketsPerIteration int
// 		iterations            int
// 		que                   PacketQueueInterface
// 		internalPacketQueue   PacketQueue
// 	}{
// 		{"Slice Queue 10 Packets 100 times", 10, 10, &PacketSliceQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Ringbuf Queue 10 Packets 100 times", 10, 10, &PacketBufQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Custom Queue 10 Packets 100 times", 10, 10, &CustomPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 		{"Channel Queue 10 Packets 100 times", 10, 10, &ChannelPacketQueue{pktQue: pq1, mutex: &sync.Mutex{}, tb: bucket}, pq1},
// 	}

// 	for _, bm := range benchmarks {
// 		b.Run(bm.name, func(b *testing.B) {
// 			bm.que.InitQueue(bm.internalPacketQueue, &sync.Mutex{}, &sync.Mutex{})
// 			var pkts []*QPkt
// 			_ = pkts
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				bm.que.Enqueue(&pkt)
// 				_ = bm.que.Pop()
// 			}
// 		})
// 	}
// }

// func benchmarkPop(popNo int, b *testing.B) {
// 	que := setupQueue()
// 	pkt := setupQueuePaket()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		j := 0
// 		for j < popNo {
// 			que.Enqueue(&pkt)
// 			j++
// 		}
// 		b.StartTimer()
// 		que.PopMultiple(popNo)
// 	}
// }

// func BenchmarkPop1(b *testing.B) { benchmarkPop(1, b) }
// func BenchmarkPop5(b *testing.B) { benchmarkPop(10, b) }

// // func TestCallPacketGen(t *testing.T) {
// // 	_ = rpkt.PrepareRtrPacketSample(t)
// // }

// // func TestBasicRoute(t *testing.T) {

// // 	r, _ := setupTestRouter(t)
// // 	r.InitQos("sample-config.yaml")

// // 	r.forwarder = r.forwardPacketTest

// // 	rp := rpkt.PrepareRtrPacketSample(t)

// // 	r.queuePacket(rp)

// // 	time.Sleep(2 * time.Second)

// // 	select {
// // 	case returnedPacket := <-packets:
// // 		if returnedPacket != rp {
// // 			t.Errorf("Returned wrong packet!")
// // 		} else {
// // 			t.Log("We got the packet back 🥳👯‍♂️👯‍♀️")
// // 		}
// // 	default:
// // 		t.Errorf("Returned no packet!")
// // 	}

// // }

// // func (r *Router) forwardPacketTest(rp *rpkt.RtrPkt) {

// // 	// defer rp.Release()

// // 	packets <- rp

// // }

// // func TestHundredPacketSliceQueue(t *testing.T) {

// // 	r, _ := setupTestRouter(t)

// // 	r.InitQos("sample-config.yaml")
// // 	r.forwarder = r.forwardPacketTest

// // 	ps := make([]*rpkt.RtrPkt, 100)

// // 	for i := 0; i < 100; i++ {
// // 		rp := rpkt.PrepareRtrPacketWithStrings(t)
// // 		r.queuePacket(rp)
// // 		ps[i] = rp
// // 	}

// // 	time.Sleep(2 * time.Second)

// // 	for i := 0; i < 100; i++ {
// // 		select {
// // 		case returnedPacket := <-packets:
// // 			if returnedPacket != ps[i] {
// // 				t.Errorf("Returned wrong packet!")
// // 			} else {
// // 				t.Log("We got the packet back 🥳👯‍♂️👯‍♀️")
// // 			}
// // 		default:
// // 			t.Errorf("Returned no packet!")
// // 		}
// // 	}
// // }

// func BenchmarkReturnAction(b *testing.B) {

// 	benchmarks := []struct {
// 		name string
// 		fu   func(polAction PoliceAction, profAction PoliceAction) PoliceAction
// 		pa1  PoliceAction
// 		pa2  PoliceAction
// 	}{
// 		{"Old", ReturnActionOld, DROPNOTIFY, DROP},
// 		{"Old", ReturnActionOld, DROP, DROPNOTIFY},
// 		{"Old", ReturnActionOld, PASS, PASS},
// 		{"New", ReturnAction, DROPNOTIFY, DROP},
// 		{"New", ReturnAction, DROP, DROPNOTIFY},
// 		{"New", ReturnAction, PASS, PASS},
// 	}

// 	benchmarksPrime := []struct {
// 		name string
// 		fu   func(polAction PrimePoliceAction, profAction PrimePoliceAction) PrimePoliceAction
// 		pa1  PrimePoliceAction
// 		pa2  PrimePoliceAction
// 	}{
// 		{"Prime", ReturnActionPrime, PrimeDROPNOTIFY, PrimeDROP},
// 		{"Prime", ReturnActionPrime, PrimeDROP, PrimeDROPNOTIFY},
// 		{"Prime", ReturnActionPrime, PrimePASS, PrimePASS},
// 	}

// 	max := int(math.Pow(10, 8))
// 	for _, bm := range benchmarks {
// 		b.Run(bm.name, func(b *testing.B) {
// 			for i := 0; i < max; i++ {
// 				bm.fu(bm.pa1, bm.pa2)
// 			}
// 		})
// 	}
// 	for _, bm := range benchmarksPrime {
// 		b.Run(bm.name, func(b *testing.B) {
// 			for i := 0; i < max; i++ {
// 				bm.fu(bm.pa1, bm.pa2)
// 			}
// 		})
// 	}
// }
