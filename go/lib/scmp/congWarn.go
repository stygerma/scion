package scmp

import "github.com/scionproto/scion/go/lib/common"

//IMPL:

type congWarn interface {
	getCongWarn()
}

type IfCongState struct {
	CurrBW int /*TODO: find way to calculate this, has to be approximated over time interval, just calculate it in the refill function.
	Add field to the tokenBucket struct that shows the latest approximation of the BW*/
	QueueLength   int
	QueueFullness float64
	IfName        common.IFIDType /*Q: Joel's policies should enforce some contracts between ISP and also be inmplemented such BRs are not overwhelmed, this means that we should
	somehow add all the IFs of this BR to this field or find a way to identify this BR such that path segments over this BR will be avoided in
	the future*/
	violation main.Violation /* Include the violation and not the whole classRule, saves space and should also give some useful info.
	classRules as defined by Joel, this would introduce big space overhead, however, there are no real incentives for an ISP to share this.
	TODO: should be exportable, i.e. write getter or tell Joel to define the type classrule by writing in uppercase */
}
