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

// func setupBufQueue() packetBufQueue {

// 	bandwidth := 0
// 	priority := 1

// 	bucket := TokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
// 	que := packetBufQueue{MaxLength: 128, MinBandwidth: priority, MaxBandWidth: priority, mutex: &sync.Mutex{}, tb: bucket}

// 	que.InitQueue(&sync.Mutex{}, &sync.Mutex{})

// 	return que

// }

// func TestBasicEnqueueBuf(t *testing.T) {
// 	que := setupBufQueue()
// 	pkt := setupQueuePaket()
// 	que.Enqueue(&pkt)
// 	length := que.GetLength()
// 	if length != 1 {
// 		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
// 	}
// 	pktb := que.Pop()

// 	if &pkt != pktb {
// 		t.Errorf("Returned wrong packet")
// 	}
// }

// func TestBasicEnqueueBufMultidequeue(t *testing.T) {
// 	que := setupBufQueue()
// 	pkt := setupQueuePaket()
// 	que.Enqueue(&pkt)
// 	length := que.GetLength()
// 	if length != 1 {
// 		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
// 	}
// 	pktb := que.PopMultiple(1)[0]

// 	if &pkt != pktb {
// 		t.Errorf("Returned wrong packet")
// 	}
// }
