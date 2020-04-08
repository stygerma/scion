package scmp

//IMPL: Defines the new SCMP type used for cumulated information dissemination

// import (
// 	"github.com/scionproto/scion/go/lib/addr"
// 	"github.com/scionproto/scion/go/lib/common"
// 	"github.com/scionproto/scion/go/lib/spath"
// )

// //defines the layout of the hop-by-hop congestion warning payload

// type InfoHbhCW struct {
// 	AsBlocks []*ASBlock
// 	RevPath  spath.Path
// }

// type ASBlock struct {
// 	SrcIA   addr.IA
// 	pktInfo []*PktInfoHbhCW
// 	mac     common.RawBytes //MS: This will be a placeholder only as DRKey is not fully functional atm
// }

// type PktInfoHbhCW struct {
// 	//RevPath   *spath.Path
// 	//Q: Do we need a SrcType or is the HostAddr sufficient?
// 	SrcHost addr.HostAddr
// 	HbhCW   *HbhCW
// }

// //var _ Info = (*HbhCW)(nil) //Interface assertion

// const hbhCWLen = 40 //everything exluding Raw of the Path TODO: increase back to 56 when spath can be included

// type HbhCW struct {
// 	CurrBW        uint64
// 	QueueLength   uint64
// 	QueueFullness uint64
// 	ConsIngress   common.IFIDType
// 	Violation     uint64
// 	//Path          *spath.Path

// 	//QueueNo       uint64 //MS: used for debugging //TODO: either fully include or remove
// 	// timestamp uint32 //TBA: maybe 64 bit needed
// 	// //ifInfo    IfCongState //MS: need to implement something such that the interface state is defined while the ISP can restrict what is shared
// 	// asBlocks []asBlock //MS: define the AS blocks
// }

//TODO: change the methods below to using HbhCW and not InfoCW which is now the information for all RAS included in one SCMP
// func InfoHbhCWFromRaw(b common.RawBytes) (*InfoHbhCW, error) {
// 	if len(b) < bscCWLen {
// 		return nil, serrors.New("Unable to parse InfoBscCW, small buffer size")
// 	}
// 	i := &InfoHbhCW{}
// 	i.CurrBW = common.Order.Uint64(b)
// 	i.QueueLength = common.Order.Uint64(b[8:])
// 	i.QueueFullness = common.Order.Uint64(b[16:])
// 	i.ConsIngress = common.IFIDType(common.Order.Uint64(b[24:]))
// 	i.Violation = common.Order.Uint64(b[32:])
// 	//.Path = spath.New(b[40:])
// 	//i.Path, _ = InfoPathFromRaw(b[40:])
// 	//i.QueueNo = common.Order.Uint64((b[24:]))
// 	return i, nil
// }

// func (i *InfoHbhCW) Copy() Info {
// 	if i == nil {
// 		return nil
// 	}
// 	return &InfoHbhCW{CurrBW: i.CurrBW, QueueLength: i.QueueLength, QueueFullness: i.QueueFullness,
// 		ConsIngress: i.ConsIngress, Violation: i.Violation} //, Path: i.Path
// }

// func (i *InfoHbhCW) Len() int {
// 	return hbhCWLen + util.CalcPadding(hbhCWLen, common.LineLen) //+ i.Path.Raw.Len()	+i.Path.Raw.Len()
// }

// func (i *InfoHbhCW) Write(b common.RawBytes) (int, error) {
// 	common.Order.PutUint64(b, i.CurrBW)
// 	common.Order.PutUint64(b[8:], i.QueueLength)
// 	common.Order.PutUint64(b[16:], i.QueueFullness)
// 	common.Order.PutUint64(b[24:], uint64(i.ConsIngress))
// 	common.Order.PutUint64(b[32:], i.Violation)
// 	//common.Order.PutUint64(b[24:], i.QueueNo)

// 	// if _, err := (i.Path.Raw).WritePld(b[40:]); err != nil {
// 	// 	return 0, err
// 	// }
// 	// if i.Path.InfOff < 0 {
// 	// 	return 0, common.NewBasicError("Negative InfOff", nil, "cannot convert to uint64")
// 	// }
// 	// common.Order.PutUint64(b[i.Path.Raw.Len()+40:], uint64(i.Path.InfOff))
// 	// if i.Path.HopOff < 0 {
// 	// 	return 0, common.NewBasicError("Negative HopOff", nil, "cannot convert to uint64")
// 	// }
// 	// common.Order.PutUint64(b[i.Path.Raw.Len()+48:], uint64(i.Path.HopOff))
// 	return util.FillPadding(b, hbhCWLen, common.LineLen), nil //i.Path.Raw.Len()
// }

// func (i *InfoHbhCW) String() string {
// 	return fmt.Sprintf("CurrBW=%d QueueLength=%d QueueFullness=%d  ConsIngress=%d Violation=%d ", //Path: Raw=%s InfOff=%d HopOff=%d		QueueNo=%d
// 		i.CurrBW, i.QueueLength, i.QueueFullness, i.ConsIngress, i.Violation) //, i.Path.Raw.String(), i.Path.InfOff, i.Path.HopOff		, i.QueueNo
// }

// const maxRascaSources = 512

// type HbhSelection struct {
// 	// lastUpdate time.Time            //starting time of the time interval
// 	// interval   time.Duration        //Length of the time interval
// 	//pkts   *ringbuf.Ring //packetinfos of sources that need to be notified for the current time interval
// 	pkts   []*PktInfoHbhCW
// 	ticker *time.Ticker
// }

// func (h *HbhSelection) Init() (timeInterval uint64) {
// 	h.ticker = time.NewTicker(time.Duration(timeInterval) * time.Millisecond)
// 	// h.pkts = ringbuf.New(maxRascaSources, func() interface{} {
// 	// 	return &PktInfoHbhCW{}
// 	// }, "idk yet")
// 	h.pkts = make([]*PktInfoHbhCW, 0, maxRascaSources)
// 	return
// }

// func (h *HbhSelection) addSelection(pktInfo *PktInfoHbhCW) {
// 	h.pkts = append(h.pkts, pktInfo)
// }
