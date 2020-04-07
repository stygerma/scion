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
	"github.com/scionproto/scion/go/lib/xtest"
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
		require.Equal(t, queueNo, tab.goldenQueueNo, "%d Queue number should be %d but is %d",
			k, tab.goldenQueueNo, queueNo)
	}
}

func TestCompareIAs(t *testing.T) {
	cases := []struct {
		a string
		b string
		r int
	}{
		// fully determined
		{a: "1-ff00:0:1", b: "1-ff00:0:2", r: -1},
		{a: "1-ff00:0:2", b: "1-ff00:0:2", r: 0},
		{a: "1-ff00:0:3", b: "1-ff00:0:2", r: +1},
		{a: "1-ff00:0:2", b: "2-ff00:0:3", r: -1},
		{a: "2-ff00:0:1", b: "1-ff00:0:2", r: +1},
		// indetermined at I or A, -1 result
		{a: "0-ff00:0:1", b: "1-ff00:0:2", r: -1},
		{a: "2-ff00:0:1", b: "0-ff00:0:2", r: -1},
		{a: "1-ff00:0:2", b: "2-0", r: -1},
		{a: "1-0", b: "2-ff00:0:2", r: -1},
		// 0 result
		{a: "1-0", b: "1-ff00:0:2", r: 0},
		{a: "0-ff00:0:1", b: "1-ff00:0:1", r: 0},
		// +1 result
		{a: "0-ff00:0:2", b: "1-ff00:0:1", r: +1},
		{a: "2-ff00:0:2", b: "0-ff00:0:1", r: +1},
		{a: "2-ff00:0:2", b: "1-0", r: +1},
		{a: "2-0", b: "1-ff00:0:2", r: +1},
		// special case
		{a: "0-0", b: "1-ff00:0:2", r: 0}, // one operand fully indetermined
	}
	for i, c := range cases {
		r := qosqueues.CompareIAs(xtest.MustParseIA(c.a), xtest.MustParseIA(c.b))
		require.Equal(t, r, c.r, "Failure at case %d", i)
		r = qosqueues.CompareIAs(xtest.MustParseIA(c.b), xtest.MustParseIA(c.a))
		require.Equal(t, r, -1*c.r, "Failure (reverse) at case %d", i)
	}
}

func forwardPacketByDrop(rp *rpkt.RtrPkt) {
	rp.Release()
}
