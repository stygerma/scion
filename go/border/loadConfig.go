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

package main

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/scionproto/scion/go/border/qosqueues"

	"github.com/scionproto/scion/go/lib/log"
	"gopkg.in/yaml.v2"
)

func (r *Router) loadConfigFile(path string) error {

	var internalRules []qosqueues.InternalClassRule
	var internalQueues []qosqueues.PacketQueueInterface

	var rc qosqueues.RouterConfig

	// dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// log.Debug("Current Path is", "path", dir)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &rc)
	if err != nil {
		return err
	}

	for _, rule := range rc.Rules {
		intRule, err := qosqueues.ConvClassRuleToInternal(rule)
		if err != nil {
			log.Error("Error reading config file", "error", err)
		}
		internalRules = append(internalRules, intRule)
	}

	for _, extQue := range rc.Queues {
		muta := &sync.Mutex{}
		mutb := &sync.Mutex{}
		queueToUse.InitQueue(que, muta, mutb)
		internalQueues = append(internalQueues, queueToUse)
	}

	r.legacyConfig = rc
	r.config = qosqueues.InternalRouterConfig{Queues: internalQueues, Rules: internalRules}

	return nil
}

func convertExternalToInteralQueue(extQueue qosqueues.ExternalPacketQueue) qosqueues.PacketQueue {

	pq := qosqueues.PacketQueue{
		Name:         extQueue.Name,
		ID:           extQueue.ID,
		MinBandwidth: extQueue.MinBandwidth,
		MaxBandWidth: extQueue.MaxBandWidth,
		PoliceRate:   convStringToNumber(extQueue.PoliceRate),
		Priority:     extQueue.Priority,
		CongWarning:  extQueue.CongWarning,
		Profile:      extQueue.Profile,
	}

	return pq
}

func convStringToNumber(bandwidthstring string) int {
	prefixes := map[string]int{
		"h": 2,
		"k": 3,
		"M": 6,
		"G": 9,
		"T": 12,
		"P": 15,
		"E": 18,
		"Z": 21,
		"Y": 24,
	}

	var num, powpow int

	for ind, str := range bandwidthstring {
		if val, contains := prefixes[string(str)]; contains {
			powpow = val
			num, _ = strToInt(bandwidthstring[:ind])
			return int(float64(num) * math.Pow(10, float64(powpow)))
		}
	}

	val, _ := strToInt(bandwidthstring)

	return val
}

func strToInt(str string) (int, error) {
	nonFractionalPart := strings.Split(str, ".")
	return strconv.Atoi(nonFractionalPart[0])
}
