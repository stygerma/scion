// Copyright 2020 ETH Zurich
// Copyright 2018 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"strings"

	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

// TODO: Matching rules is currently based on string comparisons

// Rule contains a rule for matching packets
type classRule struct {
	// This is currently means the ID of the sending border router
	Name                 string `yaml:"name"`
	Priority             int    `yaml:"priority"`
	SourceAs             string `yaml:"sourceAs"`
	SourceMatchMode      int    `yaml:"sourceMatchMode"`
	NextHopAs            string `yaml:"nextHopAs"`
	NextHopMatchMode     int    `yaml:"nextHopMatchMode"`
	DestinationAs        string `yaml:"destinationAs"`
	DestinationMatchMode int    `yaml:"destinationMatchMode"`
	L4Type               []int  `yaml:"L4Type"`
	QueueNumber          int    `yaml:"queueNumber"`
}

type internalClassRule struct {
	// This is currently means the ID of the sending border router
	Name          string
	Priority      int
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

	rule := internalClassRule{
		Name:          cr.Name,
		Priority:      cr.Priority,
		SourceAs:      sourceMatch,
		NextHopAs:     matchRule{},
		DestinationAs: destinationMatch,
		L4Type:        cr.L4Type,
		QueueNumber:   cr.QueueNumber}

	return rule, nil
}

func rulesToMap(crs []internalClassRule) (map[addr.IA][]*internalClassRule, map[addr.IA][]*internalClassRule) {
	sourceRules := make(map[addr.IA][]*internalClassRule)
	destinationRules := make(map[addr.IA][]*internalClassRule)

	for _, cr := range crs {
		if cr.SourceAs.matchMode == RANGE {
			lowLimI := uint16(cr.SourceAs.lowLim.I)
			upLimI := uint16(cr.SourceAs.upLim.I)
			lowLimA := uint64(cr.SourceAs.lowLim.A)
			upLimA := uint64(cr.SourceAs.upLim.A)

			for i := lowLimI; i <= upLimI; i++ {
				for j := lowLimA; j <= upLimA; j++ {
					sourceRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}] = append(sourceRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}], &cr)
				}
			}

		} else {

			sourceRules[cr.SourceAs.IA] = append(sourceRules[cr.SourceAs.IA], &cr)
		}
		if cr.DestinationAs.matchMode == RANGE {
			lowLimI := uint16(cr.DestinationAs.lowLim.I)
			upLimI := uint16(cr.DestinationAs.upLim.I)
			lowLimA := uint64(cr.DestinationAs.lowLim.A)
			upLimA := uint64(cr.DestinationAs.upLim.A)

			for i := lowLimI; i <= upLimI; i++ {
				for j := lowLimA; j <= upLimA; j++ {
					addr := addr.IA{I: addr.ISD(i), A: addr.AS(j)}
					destinationRules[addr] = append(destinationRules[addr], &cr)
				}
			}
		} else {
			destinationRules[cr.DestinationAs.IA] = append(destinationRules[cr.DestinationAs.IA], &cr)
		}
	}

	return sourceRules, destinationRules

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

func getQueueNumberWithHashFor(r *Router, rp *rpkt.RtrPkt) int {

	srcAddr, _ := rp.SrcIA()
	dstAddr, _ := rp.DstIA()

	queues1 := r.config.SourceRules[srcAddr]
	queues2 := r.config.DestinationRules[dstAddr]

	matches := make([]internalClassRule, 0)
	returnRule := internalClassRule{QueueNumber: 0}

	for _, rul1 := range queues1 {
		for _, rul2 := range queues2 {
			if rul1 == rul2 {
				matches = append(matches, *rul1)
			}
		}
	}

	max := -1
	for _, rul1 := range matches {
		if rul1.Priority > max {
			returnRule = rul1
			max = rul1.Priority
		}
	}

	return returnRule.QueueNumber
}

func getQueueNumberIterativeForInternal(r *Router, rp *rpkt.RtrPkt) int {

	queueNo := 0
	matches := make([]internalClassRule, 0)

	for _, cr := range r.config.Rules {

		if cr.matchInternalRule(rp) {
			matches = append(matches, cr)
		}
	}

	max := -1
	for _, rul1 := range matches {
		if rul1.Priority > max {
			queueNo = rul1.QueueNumber
			max = rul1.Priority
		}
	}

	return queueNo
}

func getQueueNumberIterativeFor(r *Router, rp *rpkt.RtrPkt) int {
	queueNo := 0

	matches := make([]classRule, 0)

	for _, cr := range r.legacyConfig.Rules {
		if cr.matchRule(rp) {
			matches = append(matches, cr)
		}
	}

	max := -1
	for _, rul1 := range matches {
		if rul1.Priority > max {
			queueNo = rul1.QueueNumber
			max = rul1.Priority
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
	destinationMatches := cr.matchSingleRule(rp, &cr.DestinationAs, rp.DstIA)
	// nextHopMatches := cr.matchSingleRule(rp, &cr.NextHopAs, rp.SrcIA)
	nextHopMatches := true

	return sourceMatches && destinationMatches && nextHopMatches
}

func (cr *classRule) matchRule(rp *rpkt.RtrPkt) bool {

	match := true

	srcAddr, _ := rp.SrcIA()
	// log.Debug("Source Address is " + srcAddr.String())
	// log.Debug("Comparing " + srcAddr.String() + " and " + cr.SourceAs)
	if !strings.Contains(srcAddr.String(), cr.SourceAs) {
		match = false
	}

	dstAddr, _ := rp.DstIA()
	// log.Debug("Destination Address is " + dstAddr.String())
	// log.Debug("Comparing " + dstAddr.String() + " and " + cr.DestinationAs)
	if !strings.Contains(dstAddr.String(), cr.DestinationAs) {
		match = false
	}

	// log.Debug("L4Type is", "L4Type", rp.CmnHdr.NextHdr)
	// log.Debug("L4Type as int is", "L4TypeInt", int(rp.CmnHdr.NextHdr))
	if !contains(cr.L4Type, int(rp.CmnHdr.NextHdr)) {
		match = false
	} else {
		// log.Debug("Matched an L4Type!")
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
