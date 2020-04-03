// Copyright 2020 ETH Zurich
// Copyright 2020 ETH Zurich, Anapaya Systems
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

package qosqueues

import (
	"strings"

	"github.com/scionproto/scion/go/border/qos/qosconf"

	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

// TODO: Matching rules is currently based on string comparisons

type InternalClassRule struct {
	// This is currently means the ID of the sending border router
	Name          string
	Priority      int
	SourceAs      matchRule
	NextHopAs     matchRule
	DestinationAs matchRule
	L4Type        []common.L4ProtocolType
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

func ConvClassRuleToInternal(cr qosconf.ExternalClassRule) (InternalClassRule, error) {

	sourceMatch, err := getMatchRuleTypeFromRule(cr, cr.SourceMatchMode, cr.SourceAs)
	if err != nil {
		return InternalClassRule{}, err
	}
	destinationMatch, err := getMatchRuleTypeFromRule(cr, cr.DestinationMatchMode, cr.DestinationAs)
	if err != nil {
		return InternalClassRule{}, err
	}

	l4t := make([]common.L4ProtocolType, 0)

	for _, l4pt := range cr.L4Type {
		l4t = append(l4t, common.L4ProtocolType(l4pt))

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

func RulesToMap(crs []InternalClassRule) *MapRules {
	sourceRules := make(map[addr.IA][]*InternalClassRule)
	destinationRules := make(map[addr.IA][]*InternalClassRule)

	asOnlySourceRules := make(map[addr.AS][]*InternalClassRule)
	asOnlyDestRules := make(map[addr.AS][]*InternalClassRule)
	isdOnlySourceRules := make(map[addr.ISD][]*InternalClassRule)
	isdOnlyDestRules := make(map[addr.ISD][]*InternalClassRule)
	sourceAnyDestinationRules := make(map[addr.IA][]*InternalClassRule)
	destinationAnySourceRules := make(map[addr.IA][]*InternalClassRule)

	for k, cr := range crs {

		switch cr.SourceAs.matchMode {
		case EXACT:
			sourceRules[cr.SourceAs.IA] = append(sourceRules[cr.SourceAs.IA], &crs[k])
		case RANGE:
			lowLimI := uint16(cr.SourceAs.lowLim.I)
			upLimI := uint16(cr.SourceAs.upLim.I)
			lowLimA := uint64(cr.SourceAs.lowLim.A)
			upLimA := uint64(cr.SourceAs.upLim.A)

			for i := lowLimI; i <= upLimI; i++ {
				for j := lowLimA; j <= upLimA; j++ {
					sourceRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}] = append(sourceRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}], &crs[k])
				}
			}
		case ASONLY:
			asOnlySourceRules[cr.SourceAs.IA.A] = append(asOnlySourceRules[cr.SourceAs.IA.A], &crs[k])
		case ISDONLY:
			isdOnlySourceRules[cr.SourceAs.IA.I] = append(isdOnlySourceRules[cr.SourceAs.IA.I], &crs[k])
		case ANY:
			destinationAnySourceRules[cr.SourceAs.IA] = append(destinationAnySourceRules[cr.SourceAs.IA], &crs[k])
		}

		switch cr.DestinationAs.matchMode {
		case EXACT:
			destinationRules[cr.DestinationAs.IA] = append(destinationRules[cr.DestinationAs.IA], &crs[k])
		case RANGE:
			lowLimI := uint16(cr.DestinationAs.lowLim.I)
			upLimI := uint16(cr.DestinationAs.upLim.I)
			lowLimA := uint64(cr.DestinationAs.lowLim.A)
			upLimA := uint64(cr.DestinationAs.upLim.A)

			//log.Debug("lowLimI", "lowLimI", lowLimI)
			//log.Debug("upLimI", "upLimI", upLimI)
			//log.Debug("lowLimA", "lowLimA", lowLimA)
			//log.Debug("upLimA", "upLimA", upLimA)

			for i := lowLimI; i <= upLimI; i++ {
				for j := lowLimA; j <= upLimA; j++ {
					//log.Debug("Adding", "I", i, "AS", j)
					destinationRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}] = append(destinationRules[addr.IA{I: addr.ISD(i), A: addr.AS(j)}], &crs[k])
				}
			}
		case ASONLY:
			asOnlyDestRules[cr.DestinationAs.IA.A] = append(asOnlyDestRules[cr.DestinationAs.IA.A], &crs[k])
		case ISDONLY:
			//log.Debug("Adding ISDONLY Destination Rule", "IA.I", cr.DestinationAs.IA.I)
			isdOnlyDestRules[cr.DestinationAs.IA.I] = append(isdOnlyDestRules[cr.DestinationAs.IA.I], &crs[k])
		case ANY:
			sourceAnyDestinationRules[cr.SourceAs.IA] = append(sourceAnyDestinationRules[cr.SourceAs.IA], &crs[k])
		}
	}

	mp := MapRules{
		RulesList:                 crs,
		SourceRules:               sourceRules,
		DestinationRules:          destinationRules,
		SourceAnyDestinationRules: sourceAnyDestinationRules,
		DestinationAnySourceRules: destinationAnySourceRules,
		ASOnlySourceRules:         asOnlySourceRules,
		ASOnlyDestRules:           asOnlyDestRules,
		ISDOnlySourceRules:        isdOnlySourceRules,
		ISDOnlyDestRules:          isdOnlyDestRules,
	}

	return &mp

}

func getMatchRuleTypeFromRule(cr qosconf.ExternalClassRule, matchModeField int, matchRuleField string) (matchRule, error) {
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
			lowLim, err := addr.IAFromString(parts[0])
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

var returnRule *InternalClassRule
var exactAndRangeSourceMatches, exactAndRangeDestinationMatches, sourceAnyDestinationMatches, destinationAnySourceRules, asOnlySourceRules, asOnlyDestinationRules []*InternalClassRule
var isdOnlySourceRules, isdOnlyDestinationRules, matched []*InternalClassRule

var sources [3][]*InternalClassRule
var destinations [3][]*InternalClassRule

var emptyRule = &InternalClassRule{
	Name:        "default",
	Priority:    0,
	QueueNumber: 0,
}

func GetRuleForPacket(config *InternalRouterConfig, rp *rpkt.RtrPkt) *InternalClassRule {

	returnRule = emptyRule

	srcAddr, _ := rp.SrcIA()
	dstAddr, _ := rp.DstIA()

	exactAndRangeSourceMatches = config.Rules.SourceRules[srcAddr]
	exactAndRangeDestinationMatches = config.Rules.DestinationRules[dstAddr]

	sourceAnyDestinationMatches = config.Rules.SourceAnyDestinationRules[srcAddr]
	destinationAnySourceRules = config.Rules.DestinationAnySourceRules[srcAddr]

	asOnlySourceRules = config.Rules.ASOnlySourceRules[srcAddr.A]
	asOnlyDestinationRules = config.Rules.ASOnlyDestRules[dstAddr.A]

	isdOnlySourceRules = config.Rules.ISDOnlySourceRules[srcAddr.I]
	isdOnlyDestinationRules = config.Rules.ISDOnlyDestRules[dstAddr.I]

	sources[0] = exactAndRangeSourceMatches
	sources[1] = asOnlySourceRules
	sources[2] = isdOnlySourceRules

	destinations[0] = exactAndRangeDestinationMatches
	destinations[1] = asOnlyDestinationRules
	destinations[2] = isdOnlyDestinationRules

	matched = intersectListsRules(sources, destinations)

	matchL4Type(&matched, rp)

	max := -1
	max, returnRule = getRuleWithPrevMax(returnRule, matched, max)
	max, returnRule = getRuleWithPrevMax(returnRule, sourceAnyDestinationMatches, max)
	max, returnRule = getRuleWithPrevMax(returnRule, destinationAnySourceRules, max)

	return returnRule
}

func matchL4Type(list *[]*InternalClassRule, rp *rpkt.RtrPkt) {

	l4h, _ := rp.L4Hdr(false)

	if l4h == nil {
		return
	}

	for i := 0; i < len(*list); i++ {
		matched := false
		for j := 0; j < len((*list)[i].L4Type); j++ {
			if (*list)[i].L4Type[j] == l4h.L4Type() {
				matched = true
				break
			}
		}
		if !matched {
			(*list)[i] = nil
		}
	}

}

func getRuleWithPrevMax(returnRule *InternalClassRule, list []*InternalClassRule, prevMax int) (int, *InternalClassRule) {

	for i := 0; i < len(list); i++ {
		if list[i] != nil {
			if list[i].Priority > prevMax {
				returnRule = list[i]
				prevMax = list[i].Priority
			}
		} else {
			break
		}
	}
	return prevMax, returnRule
}

var matches = make([]*InternalClassRule, 10)

func unionRules(a []*InternalClassRule, b []*InternalClassRule) []*InternalClassRule {

	return append(a, b...)
}

func intersectListsRules(a [3][]*InternalClassRule, b [3][]*InternalClassRule) []*InternalClassRule {
	for i := 0; i < len(matches); i++ {
		matches[i] = nil
	}
	k := 0

	for l := 0; l < 3; l++ {
		for m := 0; m < 3; m++ {
			lb := len(b[m])
			la := len(a[l])
			for i := 0; i < la; i++ {
				for j := 0; j < lb; j++ {
					if a[l][i] == b[m][j] {
						matches[k] = a[l][i]
						k++
					}
				}
			}
		}
	}
	return matches
}

func intersectRules(a []*InternalClassRule, b []*InternalClassRule) []*InternalClassRule {
	for i := 0; i < len(matches); i++ {
		matches[i] = nil
	}
	k := 0
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			if a[i] == b[j] {
				matches[k] = a[i]
				k++
			}
		}
	}
	return matches
}

func GetQueueNumberForPacket(config *InternalRouterConfig, rp *rpkt.RtrPkt) int {

	return GetRuleForPacket(config, rp).QueueNumber
}
