package main

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/scionproto/scion/go/border/brconf"
	"github.com/scionproto/scion/go/border/rctx"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/ringbuf"
	"github.com/scionproto/scion/go/lib/spkt"
	"github.com/scionproto/scion/go/lib/topology"
	"github.com/scionproto/scion/go/lib/xtest"
)

/*
Things to do:

1. Set up router with a topology
2. Create a packet

*/

func setupQueue() packetQueue {

	bandwidth := 0
	priority := 1

	bucket := tokenBucket{MaxBandWidth: bandwidth, tokens: bandwidth, lastRefill: time.Now(), mutex: &sync.Mutex{}}
	que := packetQueue{maxLength: 128, minBandwidth: priority, maxBandwidth: priority, mutex: &sync.Mutex{}, tb: bucket}

	return que

}

func setupQueuePaket() qPkt {

	return qPkt{queueNo: 0, rp: nil}
}

func TestBasicEnqueue(t *testing.T) {
	que := setupQueue()
	pkt := setupQueuePaket()
	que.enqueue(&pkt)
	length := que.getLength()
	if length != 1 {
		t.Errorf("Enqueue one packet should give length 1 gave length %d", length)
	}
}

func TestMultiEnqueue(t *testing.T) {
	que := setupQueue()
	pkt := setupQueuePaket()
	j := 15
	i := 0

	for i < j {
		que.enqueue(&pkt)
		i++
	}
	length := que.getLength()

	if length != j {
		t.Errorf("Enqueue one packet should give length %d gave length %d", j, length)
	}
}

func BenchmarkEnqueue(b *testing.B) {
	que := setupQueue()
	pkt := setupQueuePaket()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		que.enqueue(&pkt)
	}
}

func benchmarkPop(popNo int, b *testing.B) {
	que := setupQueue()
	pkt := setupQueuePaket()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		j := 0
		for j < popNo {
			que.enqueue(&pkt)
			j++
		}
		b.StartTimer()
		que.popMultiple(popNo)
	}
}

func BenchmarkPop1(b *testing.B) { benchmarkPop(1, b) }
func BenchmarkPop5(b *testing.B) { benchmarkPop(10, b) }

func TestPacketGen(t *testing.T) {
	_ = jfPrepareRtrPacketSample(t)
}

func TestThrowPanic(t *testing.T) {

	r, oldCtx := setupTestRouter(t)

}

func jfPrepareRtrPacketSample(t *testing.T) *RtrPkt {
	r := NewRtrPkt()
	r.Raw = xtest.MustReadFromFile(t, rawUdpPkt)
	// Set some other data that are required for the parsing to succeed:
	var config = &brconf.BRConf{
		IA: addr.IA{I: 1, A: 2},
		BR: &topology.BRInfo{
			IFs: map[common.IFIDType]*topology.IFInfo{5: nil},
		},
	}
	r.Ingress = addrIFPair{IfID: 5}
	// return r

	r.parseBasic()

	sampleUDPAddr := &net.UDPAddr{IP: net.IPv4(127, 1, 1, 111), Port: -1, Zone: "IPv6 scoped addressing zone"}

	r.Id = "TestPacket000"
	r.Raw = nil // We don't need contents for this anyways
	r.TimeIn = time.Now()
	r.DirFrom = -1 // I don't know what this is
	r.Ingress = addrIFPair{Dst: sampleUDPAddr,
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

// Copied from setup_test.go:

// setupTest sets up a test router. The test router is initially set up with the
// topology loaded from testdata.
func setupTestRouter(t *testing.T) (*Router, *rctx.Ctx) {
	r := initTestRouter(4)
	sockConf := brconf.SockConf{Default: PosixSock}
	// oldCtx contains the testdata topology.
	oldCtx := rctx.New(loadConfig(t))
	xtest.FailOnErr(t, r.setupNet(oldCtx, nil, sockConf))
	startSocks(oldCtx)
	return r, oldCtx
}

func initTestRouter(maxNumPosixInput int) *Router {
	// Init metrics.
	testInitOnce.Do(func() {
		// Reduce output displayed in goconvey.
		log.Root().SetHandler(log.DiscardHandler())
	})
	// The number of free packets has to be at least the number of posix
	// input routines times inputBufCnt. Otherwise they might get stuck
	// trying to prepare for reading from the connection.
	// See: https://github.com/scionproto/scion/issues/1981
	r := &Router{
		freePkts: ringbuf.New(maxNumPosixInput*inputBufCnt, func() interface{} {
			return rpkt.NewRtrPkt()
		}, "free_pkts"),
	}
	return r
}

// updateTestRouter calls setupNet on the provided router with new and old context.
// The cleanup function shall be called to free the allocated sockets.
func updateTestRouter(r *Router, newCtx, oldCtx *rctx.Ctx) func() {
	// Copy the context to make sure all sockets are closed,
	// even if socket pointers are modified in oldCtx.
	copyCtx := copyContext(oldCtx)
	err := r.setupNet(newCtx, oldCtx, brconf.SockConf{Default: PosixSock})
	SoMsg("err", err, ShouldBeNil)
	// Close all sockets to allow binding in subsequent tests.
	cleanUp := func() {
		closeAllSocks(newCtx)
		closeAllSocks(oldCtx)
		closeAllSocks(copyCtx)
	}
	return cleanUp
}

func copyContext(ctx *rctx.Ctx) *rctx.Ctx {
	c := &rctx.Ctx{}
	*c = *ctx
	c.ExtSockIn = make(map[common.IFIDType]*rctx.Sock)
	c.ExtSockOut = make(map[common.IFIDType]*rctx.Sock)
	for ifid, sock := range ctx.ExtSockIn {
		c.ExtSockIn[ifid] = sock
	}
	for ifid, sock := range ctx.ExtSockOut {
		c.ExtSockOut[ifid] = sock
	}
	return c
}

func closeAllSocks(ctx *rctx.Ctx) {
	if ctx != nil {
		ctx.LocSockIn.Stop()
		ctx.LocSockOut.Stop()
		for ifid := range ctx.ExtSockIn {
			ctx.ExtSockIn[ifid].Stop()
			ctx.ExtSockOut[ifid].Stop()
		}
	}
}

func loadConfig(t *testing.T) *brconf.BRConf {
	topo, err := topology.FromJSONFile("testdata/topology.json")
	xtest.FailOnErr(t, err)
	topoBR, ok := topo.BR("br1-ff00_0_111-1")
	if !ok {
		t.Fatal("BR ID not found")
	}
	return &brconf.BRConf{
		Topo: topo,
		IA:   topo.IA(),
		BR:   &topoBR,
	}
}
