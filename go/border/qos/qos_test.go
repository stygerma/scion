package qos

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/scionproto/scion/go/border/rpkt"
)

const configFileLocation = "/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/qos/sample-config.yaml"

func TestForMarc(t *testing.T) {

	fmt.Println("Hello test 2")

	qosConfig, _ := InitQos(configFileLocation, nil)

	fmt.Println("Config is", qosConfig)

	fmt.Println("Name is", qosConfig.config.Queues[0].GetPacketQueue().Name)
	fmt.Println("Name is", qosConfig.config.Queues[1].GetPacketQueue().Name)
	fmt.Println("Name is", qosConfig.config.Queues[2].GetPacketQueue().Name)

	fmt.Println("Profile is", qosConfig.config.Queues[0].GetPacketQueue().Profile)
	fmt.Println("Profile is", qosConfig.config.Queues[1].GetPacketQueue().Profile)
	fmt.Println("Profile is", qosConfig.config.Queues[2].GetPacketQueue().Profile)

	fmt.Println("CongWarning is", qosConfig.config.Queues[0].GetPacketQueue().CongWarning)
	fmt.Println("CongWarning is", qosConfig.config.Queues[1].GetPacketQueue().CongWarning)
	fmt.Println("CongWarning is", qosConfig.config.Queues[2].GetPacketQueue().CongWarning)

	// t.Errorf("Show Logs")
}

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

var blocks chan bool

func BenchmarkQueueSinglePacket(b *testing.B) {

	root := log15.Root()

	file, err := ioutil.TempFile("", "benchmark-log")
	if err != nil {
		b.Fatalf("Unexpected error: %v", err)
	}
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))

	qosConfig, _ := InitQos(configFileLocation, forwardPacketByDrop)
	singlePkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qosConfig.QueuePacket(singlePkt)
	}
}

func TestSingleEnqueue(t *testing.T) {

	root := log15.Root()

	file, err := ioutil.TempFile("", "benchmark-log")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))

	qosConfig, _ := InitQueueing(configFileLocation, forwardPacketByDrop)
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

	runs := 6 * 100 * 1000
	singleRun := 1024 // Should not exceed maximum queue length + capacity of notification

	qosConfig, _ := InitQueueing(configFileLocation, forwardPacketByDrop)
	arr := getPackets(singleRun)

	runtime.SetCPUProfileRate(500)
	f, err := os.Create("test/dataTestQueueSinglePacketProfile.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	fmt.Println("Array is", len(arr))

	for j := 0; j < runs; j++ {
		for _, pkt := range arr {
			// fmt.Println(k)
			qosConfig.QueuePacket(pkt)
		}

		for i := 0; i < len(arr); i++ {
			select {
			case <-testQueue:
			case <-qosConfig.notifications:
			}
		}
		if j < 11 || j%111 == 0 {
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
