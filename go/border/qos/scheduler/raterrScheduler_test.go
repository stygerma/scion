package scheduler

import (
	"log"
	"math"
	"os"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
)

func BenchmarkScheduler28191(b *testing.B) {

	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)
	qp := queues.QPkt{Rp: pkt, QueueNo: 0}

	queue1 := queues.ChannelPacketQueue{}
	queue1.InitQueue(
		queues.PacketQueue{
			MaxLength:    1024,
			MinBandwidth: 30,
			MaxBandWidth: 40},
		&sync.Mutex{},
		&sync.Mutex{})

	queue2 := queues.ChannelPacketQueue{}
	queue2.InitQueue(
		queues.PacketQueue{
			MaxLength:    1024,
			MinBandwidth: 60,
			MaxBandWidth: 80},
		&sync.Mutex{},
		&sync.Mutex{})

	for n := 0; n < b.N; n++ {
		mockSched := &RateRoundRobinScheduler{}
		mockSched.Init(
			queues.InternalRouterConfig{
				Queues: []queues.PacketQueueInterface{
					&queue1,
					&queue2}})

		go mockSched.Dequeuer(
			queues.InternalRouterConfig{
				Queues: []queues.PacketQueueInterface{
					&queue1,
					&queue2}},
			forwardPacketByDrop)

		for n := 0; n < 800; n++ {
			queue1.Enqueue(&qp)
		}
		time.Sleep(500 * time.Millisecond)
	}

}

func TestEnAndDequeuePackets(T *testing.T) {

	f, err := os.Create("cpu.out")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)
	qp := queues.QPkt{Rp: pkt, QueueNo: 0}

	queue1 := queues.ChannelPacketQueue{}
	queue1.InitQueue(
		queues.PacketQueue{
			MaxLength:    1024,
			MinBandwidth: 30,
			MaxBandWidth: 40},
		&sync.Mutex{},
		&sync.Mutex{})

	queue2 := queues.ChannelPacketQueue{}
	queue2.InitQueue(
		queues.PacketQueue{
			MaxLength:    1024,
			MinBandwidth: 60,
			MaxBandWidth: 80},
		&sync.Mutex{},
		&sync.Mutex{})

	mockSched := &RateRoundRobinScheduler{}
	mockSched.Init(
		queues.InternalRouterConfig{
			Queues: []queues.PacketQueueInterface{
				&queue1,
				&queue2}})

	j := 100

	for n := 0; n < int(math.Pow10(4)); n++ {

		for i := 0; i < j; i++ {
			queue1.Enqueue(&qp)
		}
		j = mockSched.dequeuePackets(&queue1, 100, forwardPacketByDrop, 0)
		// j = <-mockSched.jobs
		// fmt.Println("Dequeued", j)
		// fmt.Println("Iteration", n)
	}
}
