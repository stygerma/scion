package scheduler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/scionproto/scion/go/border/qos/queues"
	"github.com/scionproto/scion/go/border/rpkt"
)

func TestDequeueMM(t *testing.T) {

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

	mockSched := &MinMaxDeficitRoundRobinScheduler{}
	mockSched.Init(
		queues.InternalRouterConfig{
			Queues: []queues.PacketQueueInterface{&queue1,
				&queue2}})
	mockSched.quantumSum = 90

	fmt.Println("Before dequeue")

	pkt0 := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)
	pkt1 := rpkt.PrepareRtrPacketWithStrings("2-ff11:0:110", "17-ff00:0:112", 1)

	qpkt0 := queues.QPkt{QueueNo: 0, Act: queues.Action{}, Rp: pkt0}
	qpkt1 := queues.QPkt{QueueNo: 1, Act: queues.Action{}, Rp: pkt1}

	fmt.Println("Before dequeue")

	for i := 0; i < 1000; i++ {
		queue1.Enqueue(&qpkt0)
	}

	for i := 0; i < 66; i++ {
		queue2.Enqueue(&qpkt1)
	}

	fmt.Println("Round 1")
	fmt.Println("------------------------------")

	mockSched.Dequeue(&queue1, forwardPacketByDrop, 0)
	j := <-mockSched.jobs
	checkNoDequeued(t, 33, j)
	mockSched.Dequeue(&queue2, forwardPacketByDrop, 1)
	j = <-mockSched.jobs
	checkNoDequeued(t, 66, j)

	fmt.Println("Round 2")
	fmt.Println("------------------------------")

	mockSched.Dequeue(&queue1, forwardPacketByDrop, 0)
	j = <-mockSched.jobs
	checkNoDequeued(t, 33, j)
	mockSched.Dequeue(&queue2, forwardPacketByDrop, 1)
	j = <-mockSched.jobs
	checkNoDequeued(t, 0, j)

	fmt.Println("Round 3")
	fmt.Println("------------------------------")

	mockSched.Dequeue(&queue1, forwardPacketByDrop, 0)
	j = <-mockSched.jobs
	checkNoDequeued(t, 40, j)
	mockSched.Dequeue(&queue2, forwardPacketByDrop, 1)
	j = <-mockSched.jobs
	checkNoDequeued(t, 0, j)

	fmt.Println()
}

func TestDequeueMM2(t *testing.T) {

	queue1 := queues.ChannelPacketQueue{}
	queue1.InitQueue(
		queues.PacketQueue{
			MaxLength:    1024,
			MinBandwidth: 20,
			MaxBandWidth: 90},
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

	mockSched := &MinMaxDeficitRoundRobinScheduler{}
	mockSched.Init(
		queues.InternalRouterConfig{
			Queues: []queues.PacketQueueInterface{
				&queue1,
				&queue2}})
	mockSched.quantumSum = 90

	fmt.Println("Before dequeue")

	pkt0 := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)
	pkt1 := rpkt.PrepareRtrPacketWithStrings("2-ff11:0:110", "17-ff00:0:112", 1)

	qpkt0 := queues.QPkt{QueueNo: 0, Act: queues.Action{}, Rp: pkt0}
	qpkt1 := queues.QPkt{QueueNo: 1, Act: queues.Action{}, Rp: pkt1}

	fmt.Println("Before dequeue")

	for i := 0; i < 1000; i++ {
		queue1.Enqueue(&qpkt0)
	}

	for i := 0; i < 66; i++ {
		queue2.Enqueue(&qpkt1)
	}

	fmt.Println("Round 1")
	fmt.Println("------------------------------")

	mockSched.Dequeue(&queue1, forwardPacketByDrop, 0)
	j := <-mockSched.jobs
	checkNoDequeued(t, 22, j)
	mockSched.Dequeue(&queue2, forwardPacketByDrop, 1)
	j = <-mockSched.jobs
	checkNoDequeued(t, 66, j)

	fmt.Println("Round 2")
	fmt.Println("------------------------------")

	mockSched.Dequeue(&queue1, forwardPacketByDrop, 0)
	j = <-mockSched.jobs
	checkNoDequeued(t, 22, j)
	mockSched.Dequeue(&queue2, forwardPacketByDrop, 1)
	j = <-mockSched.jobs
	checkNoDequeued(t, 0, j)

	fmt.Println("Round 3")
	fmt.Println("------------------------------")

	mockSched.Dequeue(&queue1, forwardPacketByDrop, 0)
	j = <-mockSched.jobs
	checkNoDequeued(t, 80, j)
	mockSched.Dequeue(&queue2, forwardPacketByDrop, 1)
	j = <-mockSched.jobs
	checkNoDequeued(t, 0, j)

	fmt.Println()
}

func checkNoDequeued(t *testing.T, target int, actual int) {
	if target != actual {
		t.Errorf("Should have dequeued %d but have %d", target, actual)
	}
}

var testQueue = make(chan int, 1000)
var blockForwarder = make(chan int, 1)

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	// testQueue <- 0
	// rp.Release()
}

func forwardPacketByDropAndWait(rp *rpkt.RtrPkt) {
	<-blockForwarder
	testQueue <- 0
	rp.Release()
}

func TestSurplusMM(t *testing.T) {
	mockSched := &MinMaxDeficitRoundRobinScheduler{}
	mockSched.Init(queues.InternalRouterConfig{})
	mockSched.quantumSum = 60
	mockSched.schedulerSurplus = surplus{0, make([]int, 3), 100}

	alice := queues.ChannelPacketQueue{}
	alice.InitQueue(
		queues.PacketQueue{
			MaxLength:    100,
			MinBandwidth: 25,
			MaxBandWidth: 50},
		&sync.Mutex{},
		&sync.Mutex{})

	bob := queues.ChannelPacketQueue{}
	bob.InitQueue(
		queues.PacketQueue{
			MaxLength:    100,
			MinBandwidth: 15,
			MaxBandWidth: 25},
		&sync.Mutex{},
		&sync.Mutex{})

	charlie := queues.ChannelPacketQueue{}
	charlie.InitQueue(
		queues.PacketQueue{
			MaxLength:    100,
			MinBandwidth: 10,
			MaxBandWidth: 25},
		&sync.Mutex{},
		&sync.Mutex{})

	// Round 1

	// Alice dequeues 10
	mockSched.payIntoSurplus(&alice, 0, alice.GetMinBandwidth()-10)
	if mockSched.schedulerSurplus.Surplus != 15 {
		t.Errorf("Round 1 Alice: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Bob dequeues 10
	mockSched.payIntoSurplus(&bob, 1, bob.GetMinBandwidth()-10)
	if mockSched.schedulerSurplus.Surplus != 20 {
		t.Errorf("Round 1 Bob: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Charlie dequeues 10
	mockSched.payIntoSurplus(&charlie, 2, charlie.GetMinBandwidth()-10)
	if mockSched.schedulerSurplus.Surplus != 20 {
		t.Errorf("Round 1 Charlie: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Round 2

	fmt.Println("Round 2")

	// Alice dequeues 0
	mockSched.payIntoSurplus(&alice, 0, alice.GetMinBandwidth()-0)
	if mockSched.schedulerSurplus.Surplus != 45 {
		t.Errorf("Round 2 Alice: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Bob dequeues 0
	mockSched.payIntoSurplus(&bob, 1, bob.GetMinBandwidth()-0)
	if mockSched.schedulerSurplus.Surplus != 60 {
		t.Errorf("Round 2 Bob: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Charlie dequeues 50
	mockSched.getFromSurplus(&charlie, 2, 50)
	if mockSched.schedulerSurplus.Surplus != 45 {
		t.Errorf("Round 2 Charlie: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Round 3

	fmt.Println("Round 3")

	// Alice dequeues 0
	fmt.Println("Alice adds 25")
	mockSched.payIntoSurplus(&alice, 0, alice.GetMinBandwidth()-0)
	if mockSched.schedulerSurplus.Surplus != 70 {
		t.Errorf("Round 3 Alice: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Bob dequeues 0
	mockSched.payIntoSurplus(&bob, 1, bob.GetMinBandwidth()-0)
	if mockSched.schedulerSurplus.Surplus != 85 {
		t.Errorf("Round 3 Alice: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Charlie dequeues 0
	mockSched.payIntoSurplus(&charlie, 2, charlie.GetMinBandwidth()-0)
	if mockSched.schedulerSurplus.Surplus != 95 {
		t.Errorf("Round 3 Alice: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Round 4

	fmt.Println("Round 4")

	// Alice dequeues 0
	mockSched.payIntoSurplus(&charlie, 2, charlie.GetMinBandwidth()-0)
	if mockSched.schedulerSurplus.Surplus != 100 {
		t.Errorf("Round 4 Alice: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Bob dequeues 70
	mockSched.getFromSurplus(&bob, 2, 70)
	if mockSched.schedulerSurplus.Surplus != 90 {
		t.Errorf("Round 4 Bob: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}

	// Charlie dequeues 70
	mockSched.getFromSurplus(&charlie, 2, 70)
	if mockSched.schedulerSurplus.Surplus != 75 {
		t.Errorf("Round 4 Charlie: Incorrect surplus %d", mockSched.schedulerSurplus.Surplus)
	}
}

func TestAdjustForQuantum(t *testing.T) {

	mockSched := &MinMaxDeficitRoundRobinScheduler{}
	mockSched.Init(queues.InternalRouterConfig{})
	mockSched.quantumSum = 125

	alice := queues.ChannelPacketQueue{}
	alice.InitQueue(
		queues.PacketQueue{
			MaxLength:    100,
			MinBandwidth: 25,
			MaxBandWidth: 50,
			Priority:     5},
		&sync.Mutex{},
		&sync.Mutex{})

	bob := queues.ChannelPacketQueue{}
	bob.InitQueue(
		queues.PacketQueue{
			MaxLength:    100,
			MinBandwidth: 75,
			MaxBandWidth: 50,
			Priority:     15},
		&sync.Mutex{},
		&sync.Mutex{})

	charlie := queues.ChannelPacketQueue{}
	charlie.InitQueue(
		queues.PacketQueue{
			MaxLength:    100,
			MinBandwidth: 25,
			MaxBandWidth: 50,
			Priority:     5},
		&sync.Mutex{},
		&sync.Mutex{})

	adj := mockSched.adjustForQuantum(&alice)
	if adj != 20 {
		t.Errorf("Incorrect adjustment. Should be %d but is %d", 20, adj)
	}
	adj = mockSched.adjustForQuantum(&bob)
	if adj != 60 {
		t.Errorf("Incorrect adjustment. Should be %d but is %d", 60, adj)
	}
	adj = mockSched.adjustForQuantum(&charlie)
	if adj != 20 {
		t.Errorf("Incorrect adjustment. Should be %d but is %d", 20, adj)
	}

}
