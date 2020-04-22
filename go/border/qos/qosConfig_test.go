package qos

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/scionproto/scion/go/border/rpkt"
)

const configFileLocation = "/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/qos/sample-config.yaml"

func TestForMarc(t *testing.T) {

	fmt.Println("Hello test 2")

	qosConfig, _ := InitQueueing(configFileLocation, nil)

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

	qosConfig, _ := InitQueueing(configFileLocation, forwardPacketByDrop)
	singlePkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qosConfig.QueuePacket(singlePkt)
	}
}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}

func TestQueueSinglePacket(t *testing.T) {

	root := log15.Root()

	file, err := ioutil.TempFile("", "benchmark-log")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))

	qosConfig, _ := InitQueueing(configFileLocation, forwardPacketByDrop)
	arr := getPackets(7 * 5 * 1000 * 1000) // To get to around 20 seconds on my machine which will give a decent profile

	f, err := os.Create("sadJoelProfile.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	for _, pkt := range arr {
		qosConfig.QueuePacket(pkt)
	}

}
