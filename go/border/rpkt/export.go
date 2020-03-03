package rpkt

import (
	"net"
	"testing"
	"time"

	"github.com/scionproto/scion/go/border/brconf"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/spkt"
	"github.com/scionproto/scion/go/lib/topology"
)

//TODO: Make this be not dumb, I just want this export for tests. But for some reason it doesn't work.
func JFPrepareRtrPacketSample(t *testing.T) *RtrPkt {
	r := NewRtrPkt()
	// r.Raw = xtest.MustReadFromFile(t, rawUdpPkt)
	// Set some other data that are required for the parsing to succeed:
	var config = &brconf.BRConf{
		IA: addr.IA{I: 1, A: 2},
		BR: &topology.BRInfo{
			IFs: map[common.IFIDType]*topology.IFInfo{5: nil},
		},
	}
	r.Ingress = AddrIFPair{IfID: 5}
	// return r

	// r.parseBasic()

	sampleUDPAddr := &net.UDPAddr{IP: net.IPv4(127, 1, 1, 111), Port: -1, Zone: "IPv6 scoped addressing zone"}

	r.Id = "TestPacket000"
	r.Raw = nil // We don't need contents for this anyways
	r.TimeIn = time.Now()
	r.DirFrom = -1 // I don't know what this is
	r.Ingress = AddrIFPair{Dst: sampleUDPAddr,
		Src:  sampleUDPAddr,
		IfID: 0, IfLabel: "TODO set all of this stuff correctly"}
	r.Egress = []EgressPair{EgressPair{S: nil, Dst: sampleUDPAddr}}
	r.CmnHdr = spkt.CmnHdr{}
	r.IncrementedPath = false
	r.idxs = packetIdxs{} // packetIdxs provides offsets into a packet buffer to the start of various, we might actually need Raw
	r.dstIA = addr.IA{I: 2, A: 10}
	r.srcIA = addr.IA{I: 1, A: 10}
	r.dstHost = nil
	r.srcHost = nil
	r.infoF = nil
	r.hopF = nil
	r.ifCurr = nil
	r.ifNext = nil
	r.consDirFlag = nil
	r.HBHExt = nil
	r.E2EExt = nil
	r.L4Type = 1 // L4SCMP
	r.l4 = nil
	r.pld = nil
	r.hooks = hooks{}
	r.SCMPError = false
	// r.log.Logger = nil
	r.Ctx = rctx.New(config)
	r.refCnt = -1
	return r
}

//TODO: Make this be not dumb, I just want this export for tests. But for some reason it doesn't work.
func JFPrepareRtrPacketWith(sourceIA addr.IA, destinationIA addr.IA, L4Type common.L4ProtocolType) *RtrPkt {
	r := NewRtrPkt()
	// r.Raw = xtest.MustReadFromFile(t, rawUdpPkt)
	// Set some other data that are required for the parsing to succeed:
	var config = &brconf.BRConf{
		IA: addr.IA{I: 1, A: 2},
		BR: &topology.BRInfo{
			IFs: map[common.IFIDType]*topology.IFInfo{5: nil},
		},
	}
	r.Ingress = AddrIFPair{IfID: 5}

	sampleUDPAddr := &net.UDPAddr{IP: net.IPv4(127, 1, 1, 111), Port: -1, Zone: "IPv6 scoped addressing zone"}

	r.Id = "TestPacket000"
	r.Raw = nil // We don't need contents for this anyways
	r.TimeIn = time.Now()
	r.DirFrom = -1 // I don't know what this is
	r.Ingress = AddrIFPair{Dst: sampleUDPAddr,
		Src:  sampleUDPAddr,
		IfID: 0, IfLabel: "TODO set all of this stuff correctly"}
	r.Egress = []EgressPair{EgressPair{S: nil, Dst: sampleUDPAddr}}
	r.CmnHdr = spkt.CmnHdr{}
	r.IncrementedPath = false
	r.idxs = packetIdxs{} // packetIdxs provides offsets into a packet buffer to the start of various, we might actually need Raw
	r.srcIA = sourceIA
	r.dstIA = destinationIA
	r.dstHost = nil
	r.srcHost = nil
	r.infoF = nil
	r.hopF = nil
	r.ifCurr = nil
	r.ifNext = nil
	r.consDirFlag = nil
	r.HBHExt = nil
	r.E2EExt = nil
	r.L4Type = 1 // L4SCMP
	r.l4 = nil
	r.pld = nil
	r.hooks = hooks{}
	r.SCMPError = false
	// r.log.Logger = nil
	r.Ctx = rctx.New(config)
	r.refCnt = -1
	return r
}
