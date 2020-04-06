package qos

import (
	"io/ioutil"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/require"

	"github.com/scionproto/scion/go/border/qos/qosconf"
	"github.com/scionproto/scion/go/border/rpkt"
)

var blocks chan bool

func bBenchmarkQueueSinglePacket(b *testing.B) {
	root := log15.Root()
	file, err := ioutil.TempFile("", "benchmark-log")
	require.NoError(b, err)
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))

	extConfig, err := qosconf.LoadConfig("testdata/sample-config.yaml")
	require.NoError(b, err)
	qosConfig, _ := InitQos(extConfig, forwardPacketByDrop)
	singlePkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qosConfig.QueuePacket(singlePkt)
	}
}

// BenchmarkQueueSinglePacket measures the performance of the queue. Run with
// go test -v -run=^$ -bench=BenchmarkQueueSinglePacket ./go/border/qos/ \
//    -benchtime=20s -cpuprofile=newprofile.pprof
func BenchmarkQueueSinglePacket(t *testing.B) {
	root := log15.Root()
	file, err := ioutil.TempFile("", "benchmark-log")
	require.NoError(t, err)
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))

	extConfig, err := qosconf.LoadConfig("testdata/sample-config.yaml")
	require.NoError(t, err)
	qosConfig, _ := InitQos(extConfig, forwardPacketByDrop)
	arr := getPackets(1)

	t.ResetTimer()
	for n := 0; n < t.N; n++ {
		for _, pkt := range arr {
			qosConfig.QueuePacket(pkt)
		}
	}
}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}

func getPackets(numberOfPackets int) []*rpkt.RtrPkt {
	pkts := []*rpkt.RtrPkt{
		rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1),
		rpkt.PrepareRtrPacketWithStrings("2-ff00:0:212", "1-ff00:0:111", 1),
		rpkt.PrepareRtrPacketWithStrings("3-ff00:0:212", "1-ff00:0:111", 1),
		rpkt.PrepareRtrPacketWithStrings("4-ff00:0:212", "1-ff00:0:111", 1),
		rpkt.PrepareRtrPacketWithStrings("5-ff00:0:212", "1-ff00:0:111", 1),
		rpkt.PrepareRtrPacketWithStrings("6-ff00:0:212", "1-ff00:0:111", 1),
	}
	arr := make([]*rpkt.RtrPkt, numberOfPackets*len(pkts))
	for i := 0; i < numberOfPackets; i++ {
		copy(arr[i*len(pkts):], pkts)
	}
	return arr
}
