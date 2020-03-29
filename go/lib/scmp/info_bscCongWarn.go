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
	//EndHost addr.HostAddr MS: Router can just forward the package, i.e. does not need this info
	CurrBW uint64 /*TODO: find way to calculate this, has to be approximated over time interval, just calculate it in the refill function.
	Add field to the tokenBucket struct that shows the latest approximation of the BW*/
	QueueLength   uint64
	QueueFullness uint64
	QueueNo       uint64          //MS: used for debugging
	ConsIngress   common.IFIDType /*Q: Joel's policies should enforce some contracts between ISP and also be inmplemented such BRs are not overwhelmed, this means that we should
	somehow add all the IFs of this BR to this field or find a way to identify this BR such that path segments over this BR will be avoided in
	the future*/
	/*MS: just add the ingress interface that the packet took, this should enable an end host to find the path segments which use this interface. */
	Violation uint64 /*classRules as defined by Joel, this would introduce big space overhead, however, there are no real incentives for an ISP to share this.*/
	//ClassRule *qosqueues.InternalClassRule TODO: include this in the read from raw thingy and also the constant above
	ClassRule interface{}
}

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
