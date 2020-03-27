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
	"github.com/scionproto/scion/go/lib/log"
)

// TODO: Matching rules is currently based on string comparisons

// Rule contains a rule for matching packets
type classRule struct {
	// This is currently means the ID of the sending border router
	Name          string `yaml:"name"`
	SourceAs      string `yaml:"sourceAs"`
	NextHopAs     string `yaml:"nextHopAs"`
	DestinationAs string `yaml:"DestinationAs"`
	L4Type        []int  `yaml:"L4Type"`
	QueueNumber   int    `yaml:"queueNumber"`
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
