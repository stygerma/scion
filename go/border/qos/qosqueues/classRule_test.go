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

		queueNo := qosqueues.GetQueueNumberForPacket(qosConfig.GetConfig(), pkt)
		if queueNo != tab.goldenQueueNo {
			fmt.Println(tab.srcIA, tab.dstIA)
			t.Errorf("%d Queue number should be %d but is %d", k, tab.goldenQueueNo, queueNo)
		}
	}

}

func BenchmarkRuleMatchModes(b *testing.B) {
	extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
	qosConfig, _ := qos.InitQos(extConf, forwardPacketByDrop)

	tables := []struct {
		srcIA       string
		dstIA       string
		ruleName    string
		queueNumber int
		shouldMatch bool
	}{
		{"11-ff00:0:299", "22-ff00:0:188", "Exact - Exact", 1, true},
		{"33-ff00:0:277", "44-ff00:0:166", "Exact - ISDONLY", 2, true},
		{"33-ff00:0:277", "44-ff00:0:165", "Exact - ISDONLY", 2, true},
		{"33-ff00:0:277", "44-ff00:0:000", "Exact - ISDONLY", 2, true},
		{"55-ff00:0:055", "66-ff00:0:344", "Exact - ASONLY", 3, true},
		{"55-ff00:0:055", "12-ff00:0:344", "Exact - ASONLY", 3, true},
		{"55-ff00:0:055", "13-ff00:0:344", "Exact - ASONLY", 3, true},
		{"55-ff00:0:055", "14-ff00:0:344", "Exact - ASONLY", 3, true},
		{"77-ff00:0:233", "85-ff00:0:222", "Exact - RANGE", 4, true},
		{"77-ff00:0:233", "89-ff00:0:222", "Exact - RANGE", 4, true},
		{"2-ff00:0:011", "89-ff00:0:222", "Exact - RANGE", 4, false},
		{"2-ff00:0:011", "89-ff00:0:222", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "89-ff00:0:344", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "344-ff00:0:222", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "22-344:0:222", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "123-ff00:344:222", "Exact - ANY", 5, true},
	}

	arr := make([]rpkt.RtrPkt, len(tables))

	for k, tab := range tables {
		arr[k] = *rpkt.PrepareRtrPacketWithStrings(tab.srcIA, tab.dstIA, 1)
	}

	for i := 0; i < b.N; i++ {
		for k := 0; k < len(tables); k++ {
			rul := qosqueues.GetRuleForPacket(qosConfig.GetConfig(), &arr[k])
			_ = rul

		}
	}

}

func TestRuleMatchModes(t *testing.T) {

	extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
	qosConfig, _ := qos.InitQos(extConf, forwardPacketByDrop)

	tables := []struct {
		srcIA       string
		dstIA       string
		ruleName    string
		queueNumber int
		shouldMatch bool
	}{
		{"11-ff00:0:299", "22-ff00:0:188", "Exact - Exact", 1, true},
		{"33-ff00:0:277", "44-ff00:0:166", "Exact - ISDONLY", 2, true},
		{"33-ff00:0:277", "44-ff00:0:165", "Exact - ISDONLY", 2, true},
		{"33-ff00:0:277", "44-ff00:0:000", "Exact - ISDONLY", 2, true},
		{"55-ff00:0:055", "66-ff00:0:344", "Exact - ASONLY", 3, true},
		{"55-ff00:0:055", "12-ff00:0:344", "Exact - ASONLY", 3, true},
		{"55-ff00:0:055", "13-ff00:0:344", "Exact - ASONLY", 3, true},
		{"55-ff00:0:055", "14-ff00:0:344", "Exact - ASONLY", 3, true},
		{"77-ff00:0:233", "85-ff00:0:222", "Exact - RANGE", 4, true},
		{"77-ff00:0:233", "89-ff00:0:222", "Exact - RANGE", 4, true},
		{"2-ff00:0:011", "89-ff00:0:222", "Exact - RANGE", 4, false},
		{"2-ff00:0:011", "89-ff00:0:222", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "89-ff00:0:344", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "344-ff00:0:222", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "22-344:0:222", "Exact - ANY", 5, true},
		{"2-ff00:0:011", "123-ff00:344:222", "Exact - ANY", 5, true},
	}

	for k, tab := range tables {
		pkt := rpkt.PrepareRtrPacketWithStrings(tab.srcIA, tab.dstIA, 1)

		rul := qosqueues.GetRuleForPacket(qosConfig.GetConfig(), pkt)
		queue := qosqueues.GetQueueNumberForPacket(qosConfig.GetConfig(), pkt)

		if (rul.Name == tab.ruleName) != tab.shouldMatch {
			t.Errorf("%d should match rule %v but matches rule %v", k, tab.ruleName, rul.Name)
		}

		if (queue == tab.queueNumber) != tab.shouldMatch {
			t.Errorf("%d should match queue %v but matches queue %v", k, tab.queueNumber, queue)
		}

	}

}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}
