package main

import (
	"sync"
	"testing"
	"time"
)

func setupBufQueue() packetBufQueue {

	bandwidth := 0
	priority := 1

	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	que := packetBufQueue{MaxLength: 128, MinBandwidth: priority, MaxBandWidth: priority, mutex: &sync.Mutex{}, tb: bucket}

	que.initQueue(&sync.Mutex{}, &sync.Mutex{})

	return que

}

func TestBasicEnqueueBuf(t *testing.T) {
	que := setupBufQueue()
	pkt := setupQueuePaket()
	que.enqueue(&pkt)
	length := que.getLength()
	if length != 1 {
		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
	}
	pktb := que.pop()

	if &pkt != pktb {
		t.Errorf("Returned wrong packet")
	}
}

func TestBasicEnqueueBufMultidequeue(t *testing.T) {
	que := setupBufQueue()
	pkt := setupQueuePaket()
	que.enqueue(&pkt)
	length := que.getLength()
	if length != 1 {
		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
	}
	pktb := que.popMultiple(1)[0]

	if &pkt != pktb {
		t.Errorf("Returned wrong packet")
	}
}
