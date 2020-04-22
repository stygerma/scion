// Copyright 2020 ETH Zurich
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

package queues

import (
	"strings"

	"github.com/scionproto/scion/go/border/qos/conf"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

type InternalClassRule struct {
	// This is currently means the ID of the sending border router
	Name          string
	Priority      int
	SourceAs      matchRule
	NextHopAs     matchRule
	DestinationAs matchRule
	L4Type        map[common.L4ProtocolType]struct{}
	QueueNumber   int
}

type matchRule struct {
	IA        addr.IA
	lowLim    addr.IA // Only set if matchMode is Range
	upLim     addr.IA // Only set if matchMode is Range
	matchMode matchMode
}

type matchMode int

// modes on how to match the rules.
const (
	EXACT   matchMode = 0 // EXACT match the exact ISD and AS
	ISDONLY matchMode = 1 // ISDONLY match the ISD only
	ASONLY  matchMode = 2 // ASONLY match the AS only
	RANGE   matchMode = 3 // RANGE match AS and ISD in this range
	ANY     matchMode = 4 // ANY match anything
)

func ConvClassRuleToInternal(cr conf.ExternalClassRule) (InternalClassRule, error) {
	sourceMatch, err := getMatchFromRule(cr, cr.SourceMatchMode, cr.SourceAs)
	if err != nil {
		return InternalClassRule{}, err
	}
	destinationMatch, err := getMatchFromRule(cr, cr.DestinationMatchMode, cr.DestinationAs)
	if err != nil {
		return InternalClassRule{}, err
	}

	l4t := make(map[common.L4ProtocolType]struct{}, len(cr.L4Type))
	for _, l4pt := range cr.L4Type {
		l4t[common.L4ProtocolType(l4pt)] = struct{}{}
	}

	rule := InternalClassRule{
		Name:          cr.Name,
		Priority:      cr.Priority,
		SourceAs:      sourceMatch,
		DestinationAs: destinationMatch,
		L4Type:        l4t,
		QueueNumber:   cr.QueueNumber}

	return rule, nil
}

func RulesToMap(crs []InternalClassRule) (map[addr.IA][]*InternalClassRule,
	map[addr.IA][]*InternalClassRule) {

	sourceRules := make(map[addr.IA][]*InternalClassRule)
	destinationRules := make(map[addr.IA][]*InternalClassRule)

	for i, cr := range crs {
		if cr.SourceAs.matchMode == RANGE {
			lowLimI := uint16(cr.SourceAs.lowLim.I)
			upLimI := uint16(cr.SourceAs.upLim.I)
			lowLimA := uint64(cr.SourceAs.lowLim.A)
			upLimA := uint64(cr.SourceAs.upLim.A)

			for i := lowLimI; i <= upLimI; i++ {
				for j := lowLimA; j <= upLimA; j++ {
					sourceRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}] =
						append(sourceRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}], &crs[i])
				}
			}
		} else {
			sourceRules[cr.SourceAs.IA] = append(sourceRules[cr.SourceAs.IA], &crs[i])
		}
		if cr.DestinationAs.matchMode == RANGE {
			lowLimI := uint16(cr.DestinationAs.lowLim.I)
			upLimI := uint16(cr.DestinationAs.upLim.I)
			lowLimA := uint64(cr.DestinationAs.lowLim.A)
			upLimA := uint64(cr.DestinationAs.upLim.A)

			for i := lowLimI; i <= upLimI; i++ {
				for j := lowLimA; j <= upLimA; j++ {
					addr := addr.IA{I: addr.ISD(i), A: addr.AS(j)}
					destinationRules[addr] = append(destinationRules[addr], &crs[i])
				}
			}
		} else {
			destinationRules[cr.DestinationAs.IA] =
				append(destinationRules[cr.DestinationAs.IA], &crs[i])
		}
	}

	return sourceRules, destinationRules
}

func getMatchFromRule(cr conf.ExternalClassRule, matchModeField int, matchRuleField string) (
	matchRule, error) {

	switch matchMode(matchModeField) {
	case EXACT, ASONLY, ISDONLY, ANY:
		IA, err := addr.IAFromString(matchRuleField)
		if err != nil {
			return matchRule{}, err
		}
		m := matchRule{
			IA:        IA,
			lowLim:    addr.IA{},
			upLim:     addr.IA{},
			matchMode: matchMode(matchModeField)}
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
			m := matchRule{
				IA:        addr.IA{},
				lowLim:    lowLim,
				upLim:     upLim,
				matchMode: matchMode(matchModeField)}
			return m, nil
		}
	}

	return matchRule{}, common.NewBasicError("Invalid matchMode declared", nil,
		"matchMode", matchModeField)
}

var matches = make([]InternalClassRule, 0)
var returnRule InternalClassRule

func GetRuleWithHashFor(config *InternalRouterConfig, rp *rpkt.RtrPkt) *InternalClassRule {
	srcAddr, _ := rp.SrcIA()
	dstAddr, _ := rp.DstIA()

	queues1 := config.SourceRules[srcAddr]
	queues2 := config.DestinationRules[dstAddr]

	matches = []InternalClassRule{}
	returnRule = InternalClassRule{QueueNumber: 0}

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

	return &returnRule
}

func GetQueueNumberWithHashFor(config *InternalRouterConfig, rp *rpkt.RtrPkt) int {
	return GetRuleWithHashFor(config, rp).QueueNumber
}

func getQueueNumberIterativeForInternal(config *InternalRouterConfig, rp *rpkt.RtrPkt) int {
	queueNo := 0
	matches := make([]InternalClassRule, 0)

	for _, cr := range config.Rules {

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

func (cr *InternalClassRule) matchSingleRule(rp *rpkt.RtrPkt, matchRuleField *matchRule,
	getIA func() (addr.IA, error)) bool {
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
		if CompareIAs(matchRuleField.lowLim, addr) <= 0 &&
			CompareIAs(matchRuleField.upLim, addr) >= 0 {
			return true
		}
	}
	return false
}

func (cr *InternalClassRule) matchInternalRule(rp *rpkt.RtrPkt) bool {
	sourceMatches := cr.matchSingleRule(rp, &cr.SourceAs, rp.SrcIA)
	destinationMatches := cr.matchSingleRule(rp, &cr.DestinationAs, rp.DstIA)
	l4Matches := true
	if len(cr.L4Type) != 0 {
		contains(cr.L4Type, rp.CmnHdr.NextHdr)
	}

	return sourceMatches && destinationMatches && l4Matches
}

func contains(m map[common.L4ProtocolType]struct{}, term common.L4ProtocolType) bool {
	_, found := m[term]
	return found
}

func compareNumbers(a, b int64) int {
	if a*b == 0 || a == b {
		return 0
	} else if a < b {
		return -1
	}
	return +1
}

// CompareIAs returns -1 if a < b, 0 if a == b, and +1 if a > b
func CompareIAs(a, b addr.IA) int {
	if a.I*b.I == 0 {
		return compareNumbers(int64(a.A), int64(b.A))
	} else if a.A*b.A == 0 {
		return compareNumbers(int64(a.I), int64(b.I))
	} else {
		isd := compareNumbers(int64(a.I), int64(b.I))
		switch isd {
		case 0:
			return compareNumbers(int64(a.A), int64(b.A))
		default:
			return isd
		}
	}
}
