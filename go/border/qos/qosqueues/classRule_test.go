package qosqueues_test

import (
	"fmt"
	"testing"

	"github.com/scionproto/scion/go/border/qos/qosconf"

	"github.com/scionproto/scion/go/border/qos"
	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// TODO: Add tests for MatchModes as soon as you have decided which thing

func TestRulesWithPriority(t *testing.T) {

	tables := []struct {
		srcIA         string
		dstIA         string
		configFile    string
		goldenQueueNo int
	}{
		{"2-ff00:0:212", "1-ff00:0:110", "../testdata/priority1-config.yaml", 1},
		{"2-ff00:0:212", "1-ff00:0:111", "../testdata/priority1-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority1-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority1-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:111", "../testdata/priority1-config.yaml", 2},
		{"1-ff00:0:112", "1-ff00:0:111", "../testdata/priority1-config.yaml", 11},
		{"1-ff00:0:112", "1-ff00:0:111", "../testdata/priority2-config.yaml", 22},
		{"2-ff00:0:212", "1-ff00:0:110", "../testdata/priority2-config.yaml", 1},
		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority2-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:111", "../testdata/priority2-config.yaml", 2},
		{"2-ff00:0:212", "1-ff00:0:111", "../testdata/priority2-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority2-config.yaml", 0},
	}

	for k, tab := range tables {
		extConf, _ := qosconf.LoadConfig(tab.configFile)
		qosConfig, _ := qos.InitQos(extConf, forwardPacketByDrop)
		pkt := rpkt.PrepareRtrPacketWithStrings(tab.srcIA, tab.dstIA, 1)

		queueNo := qosqueues.GetQueueNumberWithHashFor(qosConfig.GetConfig(), pkt)
		if queueNo != tab.goldenQueueNo {
			fmt.Println(tab.srcIA, tab.dstIA)
			t.Errorf("%d Queue number should be %d but is %d", k, tab.goldenQueueNo, queueNo)
		}
	}

}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}
