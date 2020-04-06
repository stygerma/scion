package scmp

//IMPL: Defines the new SCMP type used for cumulated information dissemination

import (
	"fmt"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/spath"
	"github.com/scionproto/scion/go/lib/util"
)

//defines the layout of the hop-by-hop congestion warning payload

var _ Info = (*InfoHbhCW)(nil) //Interface assertion

type PktInfoHbhCW struct {
	RevPath   *spath.Path
	SrcIA     addr.IA
	SrcHost   addr.HostAddr
	InfoHbhCW *InfoHbhCW
}

const hbhCWLen = 40

type InfoHbhCW struct {
	CurrBW        uint64
	QueueLength   uint64
	QueueFullness uint64
	ConsIngress   common.IFIDType
	Violation     uint64
	Path 	*spath.Path

	//QueueNo       uint64 //MS: used for debugging //TODO: either fully include or remove
	// timestamp uint32 //TBA: maybe 64 bit needed
	// //ifInfo    IfCongState //MS: need to implement something such that the interface state is defined while the ISP can restrict what is shared
	// asBlocks []asBlock //MS: define the AS blocks
}

type asBlock struct {
	hosts []addr.HostAddr //MS: includes the host type and many other fields which may be unnecessary
	mac   common.RawBytes //MS: maybe done with GenerateMac from go/border/braccept/parser/parser.go or from CalcMac from go/lib/spath/hop.go
}

func (i *InfoHbhCW) Copy() Info {
	if i == nil {
		return nil
	}
	return &InfoHbhCW{CurrBW: i.CurrBW, QueueLength: i.QueueLength, QueueFullness: i.QueueFullness,
		ConsIngress: i.ConsIngress, Violation: i.Violation}
}

func (i *InfoHbhCW) Len() int {
	return hbhCWLen + util.CalcPadding(hbhCWLen, common.LineLen)
}

func (i *InfoHbhCW) Write(b common.RawBytes) (int, error) {
	common.Order.PutUint64(b, i.CurrBW)
	common.Order.PutUint64(b[8:], i.QueueLength)
	common.Order.PutUint64(b[16:], i.QueueFullness)
	common.Order.PutUint64(b[24:], uint64(i.ConsIngress))
	common.Order.PutUint64(b[32:], i.Violation)
	//common.Order.PutUint64(b[24:], i.QueueNo)

	return util.FillPadding(b, bscCWLen, common.LineLen), nil
}

func (i *InfoHbhCW) String() string {
	return fmt.Sprintf("CurrBW=%d QueueLength=%d QueueFullness=%d  ConsIngress=%d Violation=%d", //QueueNo=%d
		i.CurrBW, i.QueueLength, i.QueueFullness, i.ConsIngress, i.Violation) //, i.QueueNo
}

const maxRascaSources = 512

type HbhSelection struct {
	// lastUpdate time.Time            //starting time of the time interval
	// interval   time.Duration        //Length of the time interval
	//pkts   *ringbuf.Ring //packetinfos of sources that need to be notified for the current time interval
	pkts   []*PktInfoHbhCW
	ticker *time.Ticker
}

func (h *HbhSelection) Init() (timeInterval uint64) {
	h.ticker = time.NewTicker(time.Duration(timeInterval) * time.Millisecond)
	// h.pkts = ringbuf.New(maxRascaSources, func() interface{} {
	// 	return &PktInfoHbhCW{}
	// }, "idk yet")
	h.pkts = make([]*PktInfoHbhCW, 0, maxRascaSources)
	return
}

func (h *HbhSelection) addSelection(pktInfo *PktInfoHbhCW) {
	h.pkts = append(h.pkts, pktInfo)
}
