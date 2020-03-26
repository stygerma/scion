package qosqueues

import (
	"sync"
	"testing"
	"time"
)

func setupCustomPacketQueue(length int) customPacketQueue {

	bandwidth := 0
	priority := 1

	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	que := customPacketQueue{MaxLength: length, MinBandwidth: priority, MaxBandWidth: priority, mutex: &sync.Mutex{}, tb: bucket}

	return que

}

func TestBasicCustomEnqueue(t *testing.T) {
	que := setupCustomPacketQueue(128)
	pkt := setupQueuePaket()
	que.InitQueue(&sync.Mutex{}, &sync.Mutex{})
	que.Enqueue(&pkt)
	length := que.GetLength()
	if length != 1 {
		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
	}
}

func TestBasicCustomEnqueueDequeue(t *testing.T) {
	que := setupCustomPacketQueue(128)
	pkt := setupQueuePaket()
	que.InitQueue(&sync.Mutex{}, &sync.Mutex{})
	que.Enqueue(&pkt)
	length := que.GetLength()
	if length != 1 {
		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
	}
	pk := que.Pop()
	if pk == nil {
		t.Errorf("Returned packet is nil")
	}
	if pk != &pkt {
		t.Errorf("Wrong packet returned")
	}
}

func TestBasicCustomEnqueueRollover(t *testing.T) {
	que := setupCustomPacketQueue(32)
	pkt := setupQueuePaket()
	que.InitQueue(&sync.Mutex{}, &sync.Mutex{})

	for i := 0; i < 64; i++ {
		que.Enqueue(&pkt)
		if i >= 32-1 {
			if que.canEnqueue() {
				t.Errorf("Should not be able to enqueue %d", i)
			}
		} else {
			if !que.canEnqueue() {
				t.Errorf("Should be able to enqueue %d", i)
			}
		}
	}

	// t.Errorf("Show log")
}

func TestBasicCustomDequeueRollover(t *testing.T) {
	que := setupCustomPacketQueue(32)
	que.InitQueue(&sync.Mutex{}, &sync.Mutex{})

	for i := 0; i < 63; i++ {
		_ = que.Pop()
		if que.canDequeue() {
			t.Errorf("Should not be able to dequeue %d", i)
		}
	}

	// t.Errorf("Show log")
}
