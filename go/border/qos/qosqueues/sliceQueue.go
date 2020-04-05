package qosqueues

import (
	"math/rand"
	"sync"
)

type PacketSliceQueue struct {
	pktQue PacketQueue

	mutex *sync.Mutex

	queue  []*QPkt
	length int
	tb     tokenBucket
}

var _ PacketQueueInterface = (*PacketSliceQueue)(nil)

func (pq *PacketSliceQueue) InitQueue(que PacketQueue, mutQue *sync.Mutex, mutTb *sync.Mutex) {

	pq.pktQue = que
	pq.mutex = mutQue
	pq.length = 0
	pq.tb = tokenBucket{}
	pq.tb.Init(pq.pktQue.PoliceRate)

}

func (pq *PacketSliceQueue) Enqueue(rp *QPkt) {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.queue = append(pq.queue, rp)
	pq.length = pq.length + 1
}

func (pq *PacketSliceQueue) canDequeue() bool {

	return pq.length > 0
}

func (pq *PacketSliceQueue) GetFillLevel() int {

	return pq.length / pq.pktQue.MaxLength
}

func (pq *PacketSliceQueue) GetLength() int {

	return pq.length
}

func (pq *PacketSliceQueue) peek() *QPkt {

	return pq.queue[0]
}

func (pq *PacketSliceQueue) Pop() *QPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[0]
	pq.queue = pq.queue[1:]
	pq.length = pq.length - 1

	return pkt
}

func (pq *PacketSliceQueue) PopMultiple(number int) []*QPkt {

	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pkt := pq.queue[:number]
	pq.queue = pq.queue[number:]
	pq.length = pq.length - number

	return pkt
}

func (pq *PacketSliceQueue) CheckAction() PoliceAction {

	level := pq.GetFillLevel()

	//log.Trace("Current level is", "level", level)
	//log.Trace("Profiles are", "profiles", pq.pktQue.Profile)

	for j := len(pq.pktQue.Profile) - 1; j >= 0; j-- {
		if level >= pq.pktQue.Profile[j].FillLevel {
			//log.Trace("Matched a rule!")
			rand := rand.Intn(100)
			if rand < (pq.pktQue.Profile[j].Prob) {
				//log.Trace("Take Action!")
				return pq.pktQue.Profile[j].Action
			}
			//log.Trace("Do not take Action")

		}
	}

	return PASS
}

func (pq *PacketSliceQueue) Police(qp *QPkt) PoliceAction {
	return pq.tb.PoliceBucket(qp)
}

func (pq *PacketSliceQueue) GetMinBandwidth() int {
	return pq.pktQue.MinBandwidth
}

func (pq *PacketSliceQueue) GetMaxBandwidth() int {
	return pq.pktQue.MaxBandWidth
}

func (pq *PacketSliceQueue) GetPriority() int {
	return pq.pktQue.Priority
}

func (pq *PacketSliceQueue) GetPacketQueue() PacketQueue {
	return pq.pktQue
}
