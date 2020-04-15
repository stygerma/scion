package qosqueues

import (
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

type ClassRuleInterface interface {
	GetRuleForPacket(config *InternalRouterConfig, rp *rpkt.RtrPkt) *InternalClassRule
}

type ParallelClassRule struct {
	result []*InternalClassRule

	sources      [4][]*InternalClassRule
	destinations [4][]*InternalClassRule
}

var _ ClassRuleInterface = (*ParallelClassRule)(nil)

func (pcr *ParallelClassRule) GetRuleForPacket(config *InternalRouterConfig, rp *rpkt.RtrPkt) *InternalClassRule {

	done := make(chan bool, 3)

	var srcAddr addr.IA
	var dstAddr addr.IA
	var l4t common.L4ProtocolType

	go func(dun chan bool) {
		srcAddr, _ = rp.SrcIA()
		dun <- true
	}(done)
	go func(dun chan bool) {
		dstAddr, _ = rp.DstIA()
		dun <- true
	}(done)
	go func(dun chan bool) {

		l4h, _ := rp.L4Hdr(false)

		// log.Debug("Adresses", "srcAddr", srcAddr, "dstAddr", dstAddr, "l4t", l4t)

		if l4h == nil {
			l4t = 0
		} else {
			l4t = l4h.L4Type()
		}

		dun <- true
	}(done)

	for i := 0; i < cap(done); i++ {
		<-done
	}

	entry := cacheEntry{srcAddress: srcAddr, dstAddress: dstAddr, l4type: l4t}

	returnRule = config.Rules.CrCache.Get(entry)

	if returnRule != nil {
		return returnRule
	}

	returnRule = emptyRule

	done = make(chan bool, 8)

	// exactAndRangeSourceMatches = config.Rules.SourceRules[srcAddr]
	go pcr.getMatchFromMap(config, &config.Rules.SourceRules, srcAddr, &pcr.sources, 0, done)
	// exactAndRangeDestinationMatches = config.Rules.DestinationRules[dstAddr]
	go pcr.getMatchFromMap(config, &config.Rules.DestinationRules, dstAddr, &pcr.destinations, 0, done)

	// sourceAnyDestinationMatches = config.Rules.SourceAnyDestinationRules[srcAddr]
	go pcr.getMatchFromMap(config, &config.Rules.SourceAnyDestinationRules, srcAddr, &pcr.sources, 3, done)
	// destinationAnySourceRules = config.Rules.DestinationAnySourceRules[dstAddr]
	go pcr.getMatchFromMap(config, &config.Rules.DestinationAnySourceRules, dstAddr, &pcr.destinations, 3, done)

	// asOnlySourceRules = config.Rules.ASOnlySourceRules[srcAddr.A]
	go pcr.getMatchASFromMap(config, &config.Rules.ASOnlySourceRules, srcAddr.A, &pcr.sources, 1, done)
	// asOnlyDestinationRules = config.Rules.ASOnlyDestRules[dstAddr.A]
	go pcr.getMatchASFromMap(config, &config.Rules.ASOnlyDestRules, dstAddr.A, &pcr.destinations, 1, done)

	// isdOnlySourceRules = config.Rules.ISDOnlySourceRules[srcAddr.I]
	go pcr.getMatchISDFromMap(config, &config.Rules.ISDOnlySourceRules, srcAddr.I, &pcr.sources, 2, done)
	// isdOnlyDestinationRules = config.Rules.ISDOnlyDestRules[dstAddr.I]
	go pcr.getMatchISDFromMap(config, &config.Rules.ISDOnlyDestRules, dstAddr.I, &pcr.destinations, 2, done)

	for i := 0; i < cap(done); i++ {
		<-done
	}

	log.Debug("Matches", "pcr.sources[0]", pcr.sources[0])
	log.Debug("Matches", "pcr.destinations[0]", pcr.destinations[0])

	matched = intersectLongListsRules(pcr.sources, pcr.destinations)

	matchL4Type(&matched, l4t)

	var result [3]*InternalClassRule

	for i := 0; i < len(result); i++ {
		result[i] = emptyRule
	}

	done = make(chan bool, 3)

	go getRuleWithMaxFrom(result[0], matched, &done)
	go getRuleWithMaxFrom(result[1], pcr.sources[3], &done)
	go getRuleWithMaxFrom(result[2], pcr.destinations[3], &done)

	for i := 0; i < cap(done); i++ {
		<-done
		if result[i].Priority > returnRule.Priority {
			returnRule = result[i]
		}
	}

	// config.Rules.CrCache.Put(entry, returnRule)

	return returnRule
}

func (pcr *ParallelClassRule) getMatchISDFromMap(config *InternalRouterConfig, m *map[addr.ISD][]*InternalClassRule, address addr.ISD, result *[4][]*InternalClassRule, resultSpot int, done chan bool) {
	returnRule = emptyRule
	exactAndRangeSourceMatches = (*m)[address]
	result[resultSpot] = exactAndRangeSourceMatches
	done <- true
}

func (pcr *ParallelClassRule) getMatchASFromMap(config *InternalRouterConfig, m *map[addr.AS][]*InternalClassRule, address addr.AS, result *[4][]*InternalClassRule, resultSpot int, done chan bool) {
	returnRule = emptyRule
	exactAndRangeSourceMatches = (*m)[address]
	result[resultSpot] = exactAndRangeSourceMatches
	done <- true
}

func (pcr *ParallelClassRule) getMatchFromMap(config *InternalRouterConfig, m *map[addr.IA][]*InternalClassRule, address addr.IA, result *[4][]*InternalClassRule, resultSpot int, done chan bool) {
	returnRule = emptyRule
	exactAndRangeSourceMatches = (*m)[address]
	result[resultSpot] = exactAndRangeSourceMatches
	done <- true
}

func intersectLongListsRules(a [4][]*InternalClassRule, b [4][]*InternalClassRule) []*InternalClassRule {
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

func getRuleWithMaxFrom(result *InternalClassRule, list []*InternalClassRule, done *chan bool) {

	prevMax := -1
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
	*done <- true
}
