// Copyright 2020 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
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

package rpkt

import (
	"fmt"
	"net"
	"time"

	"github.com/scionproto/scion/go/border/brconf"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/spkt"
	"github.com/scionproto/scion/go/lib/topology"
)

func prepareRtrPacketDetailedSample(sourceIA addr.IA, destinationIA addr.IA, L4Type common.L4ProtocolType) *RtrPkt {
	r := NewRtrPkt()
	// r.Raw = xtest.MustReadFromFile(t, rawUdpPkt)
	// Set some other data that are required for the parsing to succeed:
	var config = &brconf.BRConf{
		IA: addr.IA{I: 1, A: 2},
		BR: &topology.BRInfo{
			IFs: map[common.IFIDType]*topology.IFInfo{5: nil},
		},
	}
	r.Ingress = addrIFPair{IfID: 5}
	// return r

	// r.parseBasic()

	sampleUDPAddr := &net.UDPAddr{IP: net.IPv4(127, 1, 1, 111), Port: -1, Zone: "IPv6 scoped addressing zone"}

	r.Id = "TestPacket000"
	r.Raw = nil // We don't need contents for this anyways
	r.TimeIn = time.Now()
	r.DirFrom = -1 // I don't know what this is
	r.Ingress = addrIFPair{Dst: sampleUDPAddr,
		Src:  sampleUDPAddr,
		IfID: 0, IfLabel: "TODO set all of this stuff correctly"}
	r.Egress = []EgressPair{EgressPair{S: nil, Dst: sampleUDPAddr}}
	r.CmnHdr = spkt.CmnHdr{NextHdr: common.L4ProtocolType(1)}
	r.IncrementedPath = false
	r.idxs = packetIdxs{} // packetIdxs provides offsets into a packet buffer to the start of various, we might actually need Raw
	r.dstIA = destinationIA
	r.srcIA = sourceIA
	r.dstHost = nil
	r.srcHost = nil
	r.infoF = nil
	r.hopF = nil
	r.ifCurr = nil
	r.ifNext = nil
	r.consDirFlag = nil
	r.HBHExt = nil
	r.E2EExt = nil
	r.L4Type = L4Type // L4SCMP
	r.l4 = nil
	r.pld = nil
	r.hooks = hooks{}
	r.SCMPError = false
	// r.log.Logger = nil
	r.Ctx = rctx.New(config)
	r.refCnt = -1
	return r
}

func PrepareRtrPacketWithStrings(sourceIA string, destinationIA string, L4Type int) *RtrPkt {

	srcIA, err := addr.IAFromString(sourceIA)

	if err != nil {
		fmt.Println(err)
	}

	dstIA, err := addr.IAFromString(destinationIA)

	if err != nil {
		fmt.Println(err)
	}

	pkt := prepareRtrPacketDetailedSample(
		srcIA,
		dstIA,
		common.L4ProtocolType(L4Type))
	return pkt
}
