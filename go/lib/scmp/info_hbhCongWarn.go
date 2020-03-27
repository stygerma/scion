package scmp

//IMPL: Defines the new SCMP type used for cumulated information dissemination

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/spath"
)

//defines the layout of the hop-by-hop congestion warning payload

type HbhCongWarn struct {
	rp        spath.Path
	timestamp uint32 //TBA: maybe 64 bit needed
	//ifInfo    IfCongState //MS: need to implement something such that the interface state is defined while the ISP can restrict what is shared
	asBlocks []asBlock //MS: define the AS blocks
}

type asBlock struct {
	hosts []addr.HostAddr //MS: includes the host type and many other fields which may be unnecessary
	mac   common.RawBytes //MS: maybe done with GenerateMac from go/border/braccept/parser/parser.go or from CalcMac from go/lib/spath/hop.go
}
