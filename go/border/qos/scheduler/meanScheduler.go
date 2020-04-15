package qosscheduler

import (
	"time"

	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/log"
)

// This is a deficit round robin dequeuer. Queues with higher priority will have more packets dequeued at the same time.

type Selfish struct {
	quantumSum       int
	totalLength      int
	messages         chan bool
	totalQueueLength int

	incoming   [4]int
	lastRound  [4]int
	attempted  [4]int
	total      [4]int
	iterations int

	shouldDequeue [4]int
}

var _ SchedulerInterface = (*Selfish)(nil)

func (sched *Selfish) Init(routerConfig qosqueues.InternalRouterConfig) {

	sched.quantumSum = 0
	sched.totalLength = len(routerConfig.Queues)

	for i := 0; i < len(routerConfig.Queues); i++ {
		sched.quantumSum = sched.quantumSum + routerConfig.Queues[i].GetPriority()
	}

}

func (sched *Selfish) getNoPacketsToDequeue(totalLength int, priority int, totalPriority int) int {
	// return int(math.Floor(float64(totalLength) / float64(totalPriority) * float64(priority)))
	return totalLength / totalPriority * priority
	// return priority
}

func (sched *Selfish) Dequeue(queue qosqueues.PacketQueueInterface, forwarder func(rp *rpkt.RtrPkt), queueNo int) {

	nopkts := getNoPacketsToDequeue(sched.totalQueueLength, queue.GetPriority(), sched.quantumSum)
	pktToDequeue := nopkts

	var qp *qosqueues.QPkt

	attempted[queueNo] += pktToDequeue

	for i := 0; i < pktToDequeue; i++ {
		qp = queue.Pop()
		if qp == nil {
			break
		}
		lastRound[queueNo]++
		total[queueNo]++
		forwarder(qp.Rp)
	}
}

func (sched *Selfish) Dequeuer(routerConfig qosqueues.InternalRouterConfig, forwarder func(rp *rpkt.RtrPkt)) {
	if sched.totalLength == 0 {
		panic("There are no queues to dequeue from. Please check that Init is called")
	}
	if len(routerConfig.Queues) == 3 {
		log.Debug("Priorities", "0", routerConfig.Queues[0].GetPriority(), "1", routerConfig.Queues[1].GetPriority(), "2", routerConfig.Queues[2].GetPriority())
	}
	for range time.Tick(800 * time.Microsecond) {
		sched.totalQueueLength = 0
		for i := 0; i < sched.totalLength; i++ {
			sched.totalQueueLength += routerConfig.Queues[i].GetLength()
		}
		for i := 0; i < sched.totalLength; i++ {
			sched.Dequeue(routerConfig.Queues[i], forwarder, i)
		}

		sched.showLog(routerConfig)

		// time.Sleep(80 * time.Microsecond)
	}
}

func (sched *Selfish) UpdateIncoming(queueNo int) {
	incoming[queueNo]++
}

func (sched *Selfish) showLog(routerConfig qosqueues.InternalRouterConfig) {

	iterations++
	if time.Now().Sub(t0) > time.Duration(5*time.Second) {

		var queLen [4]int
		for i := 0; i < sched.totalLength; i++ {
			queLen[i] = routerConfig.Queues[i].GetLength()
		}
		log.Debug("Last Round was", "iterations", iterations, "incoming", incoming, "lastRound", lastRound, "attempted", attempted, "total", total, "queueLen", queLen)
		for i := 0; i < len(lastRound); i++ {
			lastRound[i] = 0
		}
		for i := 0; i < len(attempted); i++ {

			attempted[i] = 0
		}
		for i := 0; i < len(incoming); i++ {
			incoming[i] = 0
		}
		t0 = time.Now()
		iterations = 0
	}

}

func (sched *Selfish) GetMessages() *chan bool {
	return &sched.messages
}
