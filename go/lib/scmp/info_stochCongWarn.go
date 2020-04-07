package scmp

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/util"
)

//IMPL: Defines the new SCMP type used for basic information dissemination

var _ Info = (*InfoStochCW)(nil) //Interface assertion

const (
	stochCWLen = 40 //everything expect for RAW of path TODO: increase back to 56 when spath can be included
)

type InfoStochCW struct {
	CurrBW        uint64
	QueueLength   uint64
	QueueFullness uint64
	ConsIngress   common.IFIDType
	Violation     uint64
	//Path          *spath.Path
	//QueueNo       uint64          //MS: used for debugging
	//ClassRule interface{}
}

func InfoStochCWFromRaw(b common.RawBytes) (*InfoStochCW, error) {
	if len(b) < stochCWLen {
		return nil, serrors.New("Unable to parse InfoBscCW, small buffer size")
	}
	i := &InfoStochCW{}
	i.CurrBW = common.Order.Uint64(b)
	i.QueueLength = common.Order.Uint64(b[8:])
	i.QueueFullness = common.Order.Uint64(b[16:])
	i.ConsIngress = common.IFIDType(common.Order.Uint64(b[24:]))
	i.Violation = common.Order.Uint64(b[32:])
	//i.Path = spath.New(b[40:])
	//i.QueueNo = common.Order.Uint64((b[24:]))

	return i, nil
}

func (i *InfoStochCW) Copy() Info {
	if i == nil {
		return nil
	}
	return &InfoStochCW{CurrBW: i.CurrBW, QueueLength: i.QueueLength,
		QueueFullness: i.QueueFullness, ConsIngress: i.ConsIngress,
		Violation: i.Violation} //, Path: i.Path		, QueueNo: i.QueueNo
}

func (i *InfoStochCW) Len() int {
	return stochCWLen + util.CalcPadding(stochCWLen, common.LineLen) // + i.Path.Raw.Len()	+i.Path.Raw.Len()
}

func (i *InfoStochCW) Write(b common.RawBytes) (int, error) {
	common.Order.PutUint64(b, i.CurrBW)
	common.Order.PutUint64(b[8:], i.QueueLength)
	common.Order.PutUint64(b[16:], i.QueueFullness)
	common.Order.PutUint64(b[24:], uint64(i.ConsIngress))
	common.Order.PutUint64(b[32:], i.Violation)
	//_, _ = i.Path.Write(b[40:])
	// if _, err := (i.Path.Raw).WritePld(b[40:i.Path.Raw.Len()]); err != nil {
	// 	return 0, err
	// }
	// if i.Path.InfOff < 0 {
	// 	return 0, common.NewBasicError("Negative InfOff", nil, "cannot convert to uint64")
	// }
	// common.Order.PutUint64(b[i.Path.Raw.Len()+40:], uint64(i.Path.InfOff))
	// if i.Path.HopOff < 0 {
	// 	return 0, common.NewBasicError("Negative HopOff", nil, "cannot convert to uint64")
	// }
	// common.Order.PutUint64(b[i.Path.Raw.Len()+48:], uint64(i.Path.HopOff))
	//	common.Order.PutUint64(b[24:], i.QueueNo)

	return util.FillPadding(b, bscCWLen, common.LineLen), nil //+i.Path.Raw.Len()
}

func (i *InfoStochCW) String() string {
	return fmt.Sprintf("CurrBW=%d QueueLength=%d QueueFullness=%d  ConsIngress=%d Violation=%d ", //Path: Raw=%s InfOff=%d HopOff=%d		QueueNo=%d
		i.CurrBW, i.QueueLength, i.QueueFullness, i.ConsIngress, i.Violation) //, i.Path.Raw.String(), i.Path.InfOff, i.Path.HopOff		, i.QueueNo, i.Path.String()
}
