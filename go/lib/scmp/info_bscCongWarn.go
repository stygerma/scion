package scmp

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/util"
)

//IMPL: Defines the new SCMP type used for basic information dissemination

var _ Info = (*InfoBscCW)(nil) //Interface assertion

const (
	bscCWLen = 48
)

type InfoBscCW struct {
	CurrBW        uint64
	QueueLength   uint64
	QueueFullness uint64
	QueueNo       uint64 //MS: used for debugging //TODO: either fully include or remove
	ConsIngress   common.IFIDType
	Violation     uint64
	//ClassRule     interface{} //TODO: include if we can limit its length
}

//TODO: integrate the  classrule to the following functions and methods
func InfoBscCWFromRaw(b common.RawBytes) (*InfoBscCW, error) {
	if len(b) < bscCWLen {
		return nil, serrors.New("Unable to parse InfoBscCW, small buffer size")
	}
	i := &InfoBscCW{}
	i.CurrBW = common.Order.Uint64(b)
	i.QueueLength = common.Order.Uint64(b[8:])
	i.QueueFullness = common.Order.Uint64(b[16:])
	i.QueueNo = common.Order.Uint64((b[24:]))
	i.ConsIngress = common.IFIDType(common.Order.Uint64(b[32:]))
	i.Violation = common.Order.Uint64(b[40:])
	return i, nil
}

func (i *InfoBscCW) Copy() Info {
	if i == nil {
		return nil
	}
	return &InfoBscCW{CurrBW: i.CurrBW, QueueLength: i.QueueLength, QueueFullness: i.QueueFullness, QueueNo: i.QueueNo, ConsIngress: i.ConsIngress, Violation: i.Violation}
}

func (i *InfoBscCW) Len() int {
	return bscCWLen + util.CalcPadding(bscCWLen, common.LineLen)
}

func (i *InfoBscCW) Write(b common.RawBytes) (int, error) {
	common.Order.PutUint64(b, i.CurrBW)
	common.Order.PutUint64(b[8:], i.QueueLength)
	common.Order.PutUint64(b[16:], i.QueueFullness)
	common.Order.PutUint64(b[24:], i.QueueNo)
	common.Order.PutUint64(b[32:], uint64(i.ConsIngress))
	common.Order.PutUint64(b[40:], i.Violation)
	return util.FillPadding(b, bscCWLen, common.LineLen), nil
}

func (i *InfoBscCW) String() string {
	return fmt.Sprintf("CurrBW=%d QueueLength=%d QueueFullness=%d QueueNo=%d ConsIngress=%d Violation=%d",
		i.CurrBW, i.QueueLength, i.QueueFullness, i.QueueNo, i.ConsIngress, i.Violation)
}
