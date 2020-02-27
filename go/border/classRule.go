package main

import (
	"strings"

	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/log"
)

// TODO: Matching rules is currently based on string comparisons

// Rule contains a rule for matching packets
type classRule struct {
	// This is currently means the ID of the sending border router
	sourceAs      string
	nextHopAs     string
	destinationAs string
	queueNumber   int
}

func getQueueNumberFor(rp *rpkt.RtrPkt, crs *[]classRule) int {

	queueNo := 0

	srcIA, _ := rp.SrcIA()
	dstIA, _ := rp.DstIA()
	if (dstIA == addr.IA{I: 2, A: 10} &&
		srcIA == addr.IA{I: 1, A: 10}) {
		// panic(errors.New("We have received the testpacket! 🥳"))
	}

	for _, cr := range *crs {
		if cr.matchRule(rp) {
			queueNo = cr.queueNumber
		}
	}
	return queueNo
}

func (cr *classRule) matchRule(rp *rpkt.RtrPkt) bool {

	match := true

	srcAddr, _ := rp.SrcIA()
	log.Debug("Source Address is " + srcAddr.String())
	log.Debug("Comparing " + srcAddr.String() + " and " + cr.sourceAs)
	if !strings.Contains(srcAddr.String(), cr.sourceAs) {
		match = false
	}

	dstAddr, _ := rp.DstIA()
	log.Debug("Destination Address is " + dstAddr.String())
	log.Debug("Comparing " + dstAddr.String() + " and " + cr.destinationAs)
	if !strings.Contains(dstAddr.String(), cr.destinationAs) {
		match = false
	}

	return match
}
