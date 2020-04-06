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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scionproto/scion/go/border/qos"
	"github.com/scionproto/scion/go/border/qos/qosconf"
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
		{"2-ff00:0:212", "1-ff00:0:110", "testdata/priority1-config.yaml", 1},
		{"2-ff00:0:212", "1-ff00:0:111", "testdata/priority1-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:110", "testdata/priority1-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:110", "testdata/priority1-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:111", "testdata/priority1-config.yaml", 2},
		{"1-ff00:0:112", "1-ff00:0:111", "testdata/priority1-config.yaml", 11},
		{"1-ff00:0:112", "1-ff00:0:111", "testdata/priority2-config.yaml", 22},
		{"2-ff00:0:212", "1-ff00:0:110", "testdata/priority2-config.yaml", 1},
		{"1-ff00:0:110", "1-ff00:0:110", "testdata/priority2-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:111", "testdata/priority2-config.yaml", 2},
		{"2-ff00:0:212", "1-ff00:0:111", "testdata/priority2-config.yaml", 0},
		{"1-ff00:0:110", "1-ff00:0:110", "testdata/priority2-config.yaml", 0},
	}
	for k, tab := range tables {
		extConfig, err := qosconf.LoadConfig(tab.configFile)
		require.NoError(t, err, "Failed at case %d", k)
		qosConfig, err := qos.InitQos(extConfig, forwardPacketByDrop)
		require.NoError(t, err, "Failed at case %d", k)
		pkt := rpkt.PrepareRtrPacketWithStrings(tab.srcIA, tab.dstIA, 1)

		queueNo := qosqueues.GetQueueNumberWithHashFor(qosConfig.GetConfig(), pkt)
		if queueNo != tab.goldenQueueNo {
			require.Equal(t, queueNo, tab.goldenQueueNo, "%d Queue number should be %d but is %d",
				k, tab.goldenQueueNo, queueNo)
		}
	}

}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}
