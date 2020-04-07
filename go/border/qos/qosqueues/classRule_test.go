// Copyright 2020 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qosqueues_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/scionproto/scion/go/border/qos/qosconf"
	"github.com/scionproto/scion/go/lib/log"

	"github.com/scionproto/scion/go/border/qos"
	"github.com/scionproto/scion/go/border/qos/qosconf"
	"github.com/scionproto/scion/go/border/qos/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
)

// TODO: Add tests for MatchModes as soon as you have decided which thing

// func TestRulesWithPriority(t *testing.T) {

// 	tables := []struct {
// 		srcIA         string
// 		dstIA         string
// 		configFile    string
// 		goldenQueueNo int
// 	}{
// 		{"2-ff00:0:212", "1-ff00:0:110", "../testdata/priority1-config.yaml", 1},
// 		{"2-ff00:0:212", "1-ff00:0:111", "../testdata/priority1-config.yaml", 0},
// 		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority1-config.yaml", 0},
// 		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority1-config.yaml", 0},
// 		{"1-ff00:0:110", "1-ff00:0:111", "../testdata/priority1-config.yaml", 2},
// 		{"1-ff00:0:112", "1-ff00:0:111", "../testdata/priority1-config.yaml", 11},
// 		{"1-ff00:0:112", "1-ff00:0:111", "../testdata/priority2-config.yaml", 22},
// 		{"2-ff00:0:212", "1-ff00:0:110", "../testdata/priority2-config.yaml", 1},
// 		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority2-config.yaml", 0},
// 		{"1-ff00:0:110", "1-ff00:0:111", "../testdata/priority2-config.yaml", 2},
// 		{"2-ff00:0:212", "1-ff00:0:111", "../testdata/priority2-config.yaml", 0},
// 		{"1-ff00:0:110", "1-ff00:0:110", "../testdata/priority2-config.yaml", 0},
// 	}

// 	for k, tab := range tables {
// 		extConf, _ := qosconf.LoadConfig(tab.configFile)
// 		qosConfig, _ := qos.InitQos(extConf, forwardPacketByDrop)
// 		pkt := rpkt.PrepareRtrPacketWithStrings(tab.srcIA, tab.dstIA, 1)

// 		queueNo := qosqueues.GetQueueNumberForPacket(qosConfig.GetConfig(), pkt)
// 		if queueNo != tab.goldenQueueNo {
// 			fmt.Println(tab.srcIA, tab.dstIA)
// 			t.Errorf("%d Queue number should be %d but is %d", k, tab.goldenQueueNo, queueNo)
// 		}
// 	}

// }

func BenchmarkRuleMatchModes(b *testing.B) {
	// extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
	extConf, _ := qosconf.LoadConfig("../testdata/matchBenchmark-config.yaml")
	qosConfig, _ := qos.InitQos(extConf, forwardPacketByDrop)

	rc := qosqueues.RegularClassRule{}

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < len(tables); k++ {
			rul := rc.GetRuleForPacket(qosConfig.GetConfig(), &arr[k])
			_ = rul

		}
	}

}

func BenchmarkSingleMatchSequential(b *testing.B) {
	disableLog(b)
	// extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
	extConf, _ := qosconf.LoadConfig("../testdata/matchBenchmark-config.yaml")
	// qosConfig, _ := qos.InitQos(extConf, forwardPacketByDropAndWait)

	qConfig := qos.QosConfiguration{}

	var err error
	if err = qos.ConvertExternalToInternalConfig(&qConfig, extConf); err != nil {
		log.Error("Initialising the classification data structures has failed", "error", err)
	}
	if err = qos.InitClassification(&qConfig); err != nil {
		log.Error("Initialising the classification data structures has failed", "error", err)
	}

	qosConfig := qConfig

	rc := qosqueues.RegularClassRule{}

	pkt := rpkt.PrepareRtrPacketWithStrings("11-ff00:0:299", "22-ff00:0:188", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rul := rc.GetRuleForPacket(qosConfig.GetConfig(), pkt)
		_ = rul
	}
}

func BenchmarkSingleMatchParallel(b *testing.B) {
	disableLog(b)
	// extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
	extConf, _ := qosconf.LoadConfig("../testdata/matchBenchmark-config.yaml")
	// qosConfig, _ := qos.InitQos(extConf, forwardPacketByDropAndWait)

	qConfig := qos.QosConfiguration{}

	var err error
	if err = qos.ConvertExternalToInternalConfig(&qConfig, extConf); err != nil {
		log.Error("Initialising the classification data structures has failed", "error", err)
	}
	if err = qos.InitClassification(&qConfig); err != nil {
		log.Error("Initialising the classification data structures has failed", "error", err)
	}

	qosConfig := qConfig

	rc := qosqueues.ParallelClassRule{}

	pkt := rpkt.PrepareRtrPacketWithStrings("11-ff00:0:299", "22-ff00:0:188", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rul := rc.GetRuleForPacket(qosConfig.GetConfig(), pkt)
		_ = rul
		// fmt.Println("Iteration", i)
	}
}

// func BenchmarkSingleMatchSemiParallel(b *testing.B) {
// 	disableLog(b)
// 	// extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
// 	extConf, _ := qosconf.LoadConfig("../testdata/matchBenchmark-config.yaml")
// 	// qosConfig, _ := qos.InitQos(extConf, forwardPacketByDropAndWait)

// 	qConfig := qos.QosConfiguration{}

// 	var err error
// 	if err = qos.ConvertExternalToInternalConfig(&qConfig, extConf); err != nil {
// 		log.Error("Initialising the classification data structures has failed", "error", err)
// 	}
// 	if err = qos.InitClassification(&qConfig); err != nil {
// 		log.Error("Initialising the classification data structures has failed", "error", err)
// 	}

// 	qosConfig := qConfig

// 	rc := qosqueues.SemiParallelClassRule{}

// 	pkt := rpkt.PrepareRtrPacketWithStrings("11-ff00:0:299", "22-ff00:0:188", 1)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		rul := rc.GetRuleForPacket(qosConfig.GetConfig(), pkt)
// 		_ = rul
// 		// fmt.Println("Iteration", i)
// 	}
// }

func TestRuleMatchModes(t *testing.T) {

	extConf, _ := qosconf.LoadConfig("../testdata/matchTypeTest-config.yaml")
	qosConfig, _ := qos.InitQos(extConf, forwardPacketByDrop)

	rc := qosqueues.RegularClassRule{}
	rcp := qosqueues.ParallelClassRule{}
	rcsp := qosqueues.SemiParallelClassRule{}

	classifiers := [3]qosqueues.ClassRuleInterface{&rc, &rcp, &rcsp}

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

			rul := classifier.GetRuleForPacket(qosConfig.GetConfig(), pkt)
			// queue := qosqueues.GetQueueNumberForPacket(qosConfig.GetConfig(), pkt)

			if rul == nil {
				fmt.Println("Rule was nil")
			}

			if (rul.Name == tab.ruleName) != tab.shouldMatch {
				t.Errorf("%d should match rule %v %v but matches rule %v", k, tab.shouldMatch, tab.ruleName, rul.Name)
			}

			// if (queue == tab.queueNumber) != tab.shouldMatch {
			// 	t.Errorf("%d should match queue %v but matches queue %v", k, tab.queueNumber, queue)
			// }

		}
	}

}

var forward = make(chan bool, 1)

func forwardPacketByDropAndWait(rp *rpkt.RtrPkt) {
	forward <- true
	rp.Release()
}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}

func disableLog(b *testing.B) {
	root := log15.Root()

	file, err := ioutil.TempFile("", "benchmark-log")
	if err != nil {
		b.Fatalf("Unexpected error: %v", err)
	}
	root.SetHandler(log15.Must.FileHandler(file.Name(), log15.LogfmtFormat()))
}
