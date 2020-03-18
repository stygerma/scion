package main

import (
	"strings"

	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

// TODO: Matching rules is currently based on string comparisons

// Rule contains a rule for matching packets
type classRule struct {
	// This is currently means the ID of the sending border router
	Name                 string `yaml:"name"`
	SourceAs             string `yaml:"sourceAs"`
	SourceMatchMode      int    `yaml:"sourceMatchMode"`
	NextHopAs            string `yaml:"nextHopAs"`
	NextHopMatchMode     int    `yaml:"nextHopMatchMode"`
	DestinationAs        string `yaml:"DestinationAs"`
	DestinationMatchMode int    `yaml:"destinationMatchMode"`
	L4Type               []int  `yaml:"L4Type"`
	QueueNumber          int    `yaml:"queueNumber"`
}

type internalClassRule struct {
	// This is currently means the ID of the sending border router
	Name          string
	SourceAs      matchRule
	NextHopAs     matchRule
	DestinationAs matchRule
	L4Type        []int
	QueueNumber   int
}

type matchRule struct {
	IA        addr.IA
	lowLim    addr.IA // Only set if matchMode is Range
	upLim     addr.IA // Only set if matchMode is Range
	matchMode matchMode
}

type matchMode int

const (
	// EXACT match the exact ISD and AS
	EXACT matchMode = 0
	// ISDONLY match the ISD only
	ISDONLY matchMode = 1
	// ASONLY match the AS only
	ASONLY matchMode = 2
	// RANGE match AS and ISD in this range
	RANGE matchMode = 3
	// ANY match anything
	ANY matchMode = 4
)

func convClassRuleToInternal(cr classRule) (internalClassRule, error) {

	sourceMatch, err := getMatchFromRule(cr, cr.SourceMatchMode, cr.SourceAs)
	if err != nil {
		return internalClassRule{}, err
	}
	destinationMatch, err := getMatchFromRule(cr, cr.DestinationMatchMode, cr.DestinationAs)
	if err != nil {
		return internalClassRule{}, err
	}
	nextHopMatch, err := getMatchFromRule(cr, cr.NextHopMatchMode, cr.NextHopAs)
	if err != nil {
		return internalClassRule{}, err
	}

	return internalClassRule{
		Name:          cr.Name,
		SourceAs:      sourceMatch,
		NextHopAs:     nextHopMatch,
		DestinationAs: destinationMatch,
		L4Type:        cr.L4Type,
		QueueNumber:   cr.QueueNumber}, nil
}

func getMatchFromRule(cr classRule, matchModeField int, matchRuleField string) (matchRule, error) {
	switch matchMode(matchModeField) {
	case EXACT, ASONLY, ISDONLY, ANY:
		IA, err := addr.IAFromString(matchRuleField)
		if err != nil {
			return matchRule{}, err
		}
		m := matchRule{IA: IA, lowLim: addr.IA{}, upLim: addr.IA{}, matchMode: matchMode(matchModeField)}
		return m, nil
	case RANGE:
		if matchMode(matchModeField) == RANGE {
			parts := strings.Split(matchRuleField, "||")
			if len(parts) != 2 {
				return matchRule{}, common.NewBasicError("Invalid Class", nil, "raw", matchModeField)
			}
			lowLim, err := addr.IAFromString(parts[1])
			if err != nil {
				return matchRule{}, err
			}
			upLim, err := addr.IAFromString(parts[1])
			if err != nil {
				return matchRule{}, err
			}
			m := matchRule{IA: addr.IA{}, lowLim: lowLim, upLim: upLim, matchMode: matchMode(matchModeField)}
			return m, nil
		}
	}

	return matchRule{}, common.NewBasicError("Invalid matchMode declared", nil, "matchMode", matchModeField)
}

func getQueueNumberFor(rp *rpkt.RtrPkt, crs *[]classRule) int {

	queueNo := 0

	for _, cr := range *crs {
		if cr.matchRule(rp) {
			queueNo = cr.QueueNumber
		}
	}
	return queueNo
}

func (cr *internalClassRule) matchSingleRule(rp *rpkt.RtrPkt, matchRuleField *matchRule, getIA func() (addr.IA, error)) bool {

	switch matchRuleField.matchMode {
	case EXACT, ASONLY, ISDONLY, ANY:
		Addr, err := getIA()
		if err != nil {
			return false
		}
		return (*matchRuleField).IA.Equal(Addr)
	case RANGE:
		addr, err := getIA()
		if err != nil {
			return false
		}
		if addr.BiggerThan(matchRuleField.lowLim) && addr.SmallerThan(matchRuleField.upLim) {
			return true
		}
	}
	return false
}

func (cr *internalClassRule) matchInternalRule(rp *rpkt.RtrPkt) bool {

	sourceMatches := cr.matchSingleRule(rp, &cr.SourceAs, rp.SrcIA)
	destinationMatches := cr.matchSingleRule(rp, &cr.SourceAs, rp.SrcIA)
	nextHopMatches := cr.matchSingleRule(rp, &cr.SourceAs, rp.SrcIA)

	return sourceMatches && destinationMatches && nextHopMatches
}

func (cr *classRule) matchRule(rp *rpkt.RtrPkt) bool {

	match := true

	srcAddr, _ := rp.SrcIA()
	log.Debug("Source Address is " + srcAddr.String())
	log.Debug("Comparing " + srcAddr.String() + " and " + cr.SourceAs)
	if !strings.Contains(srcAddr.String(), cr.SourceAs) {
		match = false
	}

	dstAddr, _ := rp.DstIA()
	log.Debug("Destination Address is " + dstAddr.String())
	log.Debug("Comparing " + dstAddr.String() + " and " + cr.DestinationAs)
	if !strings.Contains(dstAddr.String(), cr.DestinationAs) {
		match = false
	}

	log.Debug("L4Type is", "L4Type", rp.CmnHdr.NextHdr)
	log.Debug("L4Type as int is", "L4TypeInt", int(rp.CmnHdr.NextHdr))
	if !contains(cr.L4Type, int(rp.CmnHdr.NextHdr)) {
		match = false
	} else {
		log.Debug("Matched an L4Type!")
	}

	return match
}

func contains(slice []int, term int) bool {
	for _, item := range slice {
		if item == term {
			return true
		}
	}
	return false
}
