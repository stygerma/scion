package qos

import (
	"io/ioutil"
<<<<<<< ea68ea1a78c46fe5b893e5764b4f05d8e74e1769
	"math/rand"
=======
>>>>>>> fix UT/benchmark
	"testing"
	"time"

	"github.com/inconshreveable/log15"
<<<<<<< ea68ea1a78c46fe5b893e5764b4f05d8e74e1769
=======
	"github.com/stretchr/testify/require"

>>>>>>> fix UT/benchmark
	"github.com/scionproto/scion/go/border/qos/qosconf"
	"github.com/scionproto/scion/go/border/rpkt"
)

<<<<<<< ea68ea1a78c46fe5b893e5764b4f05d8e74e1769
func getPackets(numberOfPackets int) []*rpkt.RtrPkt {

	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)
	pkt1 := rpkt.PrepareRtrPacketWithStrings("2-ff00:0:212", "1-ff00:0:111", 1)
	pkt2 := rpkt.PrepareRtrPacketWithStrings("3-ff00:0:212", "1-ff00:0:111", 1)
	pkt3 := rpkt.PrepareRtrPacketWithStrings("4-ff00:0:212", "1-ff00:0:111", 1)
	pkt4 := rpkt.PrepareRtrPacketWithStrings("5-ff00:0:212", "1-ff00:0:111", 1)
	pkt5 := rpkt.PrepareRtrPacketWithStrings("6-ff00:0:212", "1-ff00:0:111", 1)

	arr := make([]*rpkt.RtrPkt, numberOfPackets)

	for i := 0; i < numberOfPackets-5; i++ {
		arr[i+0] = pkt
		arr[i+1] = pkt1
		arr[i+2] = pkt2
		arr[i+3] = pkt3
		arr[i+4] = pkt4
		arr[i+5] = pkt5
	}

	return arr
}

=======
>>>>>>> fix UT/benchmark
var blocks chan bool

func bBenchmarkQueueSinglePacket(b *testing.B) {
	root := log15.Root()
	file, err := ioutil.TempFile("", "benchmark-log")
	require.NoError(b, err)
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))

	extConf, _ := qosconf.LoadConfig("testdata/matchBenchmark-config.yaml")
	qosConfig, _ := InitQos(extConf, forwardPacketByDrop)
	singlePkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qosConfig.QueuePacket(singlePkt)
	}
}

func TestSingleEnqueue(t *testing.T) {

	extConf, _ := qosconf.LoadConfig("testdata/matchBenchmark-config.yaml")
	qosConfig, _ := InitQos(extConf, forwardPacketByDrop)
	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

	qosConfig.QueuePacket(pkt)

}

func BenchmarkGolangRandom(b *testing.B) {

	for n := 0; n < b.N; n++ {
		_ = rand.Intn(100)
	}

}

var testQueue = make(chan int, 1000)

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	testQueue <- 0
	rp.Release()
}

func TestEnqueueWithProfile(t *testing.T) {

	start := time.Now()

	runs := 6 * 1000
	singleRun := 1000 // Should not exceed maximum queue length + capacity of notification

	extConf, _ := qosconf.LoadConfig("testdata/matchBenchmark-config.yaml")
	qosConfig, _ := InitQos(extConf, forwardPacketByDrop)

	arr := getPackets(singleRun)

	fmt.Println("Array is", len(arr))

	for j := 0; j < runs; j++ {
		for _, pkt := range arr {
			// fmt.Println("Enqueue", k)
			qosConfig.QueuePacket(pkt)
		}
		for i := 0; i < len(arr); i++ {
			// fmt.Println("Dequeue", i)
			select {
			case <-testQueue:
			case <-qosConfig.notifications:
			}
		}
		if j < 11 || j%20 == 0 {
			printLog("Runs", j, runs, start)
		}
	}

}

func printLog(leading string, j int, runs int, start time.Time) {
	if j > 0 {
		ts := time.Since(start)
		et := time.Since(start) * time.Duration(runs) / time.Duration(j)
		_ = et - ts
		// fmt.Println("Run", j, "/", runs, "in", ts.Truncate(time.Second).String(), "estimated total time", et.Truncate(time.Second).String(), "remaining", (et - ts).Truncate(time.Second).String())

		fmt.Printf("%v %06d / %06d in %v. Estimated total time %v. Remaining %v\n", leading, j, runs, ts.Truncate(time.Second).String(), et.Truncate(time.Second).String(), (et - ts).Truncate(time.Second).String())
	}
}

func BenchmarkEnqueueForProfile(b *testing.B) {

	disableLog(b)

	singleRun := 1024 // Should not exceed maximum queue length + capacity of notification

	extConf, _ := qosconf.LoadConfig("testdata/matchBenchmark-config.yaml")
	qosConfig, _ := InitQos(extConf, forwardPacketByDrop)
	arr := getPackets(singleRun)
	la := len(arr)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		for k := 0; k < la; k++ {
			qosConfig.QueuePacket(arr[k])
		}

		for i := 0; i < la; i++ {
			select {
			case <-testQueue:
			case <-qosConfig.notifications:
			}
		}
	}
}

func disableLog(b *testing.B) {
	root := log15.Root()

	file, err := ioutil.TempFile("", "benchmark-log")
	if err != nil {
		b.Fatalf("Unexpected error: %v", err)
	}
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))
}
