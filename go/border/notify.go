package main

import (
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/scmp"
)

type notify interface {
	createNotificationSCMP(qp *qPkt, bscCongWarn *scmp.BscCongWarn) (*rpkt.RtrPkt, error)
	createPld() *scmp.Payload //TODO: think about these parameters
}
