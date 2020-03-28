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
	"fmt"
	"testing"
)

func TestLoadSampleConfig(t *testing.T) {
	r, _ := setupTestRouter(t)

	r.loadConfigFile("sample-config.yaml")

	fmt.Println("The config is: ", r.config)

	fmt.Println("The Queue is: ", r.config.Queues[0])
	fmt.Println("The Rule is: ", r.config.Rules[0])
	// t.Errorf("Output: %v", r.config)

}

// TODO: Readd this test with the new qosqueues
// func TestLoadSampleConfigQueues(t *testing.T) {
// 	r, _ := setupTestRouter(t)

// 	r.loadConfigFile("sample-config.yaml")

// 	fmt.Println("The Queue is: ", r.config.Queues[0])
// 	fmt.Println("The Rule is: ", r.config.Rules[0])
// 	fmt.Println("We have this number of queues: ", len(r.config.Queues))
// 	t.Errorf("Output: %v", r.config)

	if r.config.Queues[0].GetPacketQueue().ID != 0 {
		t.Errorf("Incorrect Queue ID")
	}
	if r.config.Queues[1].GetPacketQueue().ID != 1 {
		t.Errorf("Incorrect Queue ID")
	}

	if r.config.Queues[0].GetPacketQueue().Name != "General Queue" {
		t.Errorf("Incorrect Queue Name is %v but should be %v", r.config.Queues[0].GetPacketQueue().Name, "General Queue")
	}
	if r.config.Queues[1].GetPacketQueue().Name != "Speedy Queue" {
		t.Errorf("Incorrect Queue Name is %v but should be %v", r.config.Queues[0].GetPacketQueue().Name, "Speedy Queue")
	}
}

// TODO: Move these tests somewhere else

// func TestMaps(t *testing.T) {
// 	m := make(map[addr.IA]*qosqueues.InternalClassRule)

// 	IA1, _ := addr.IAFromString("1-ff00:0:110")
// 	IA2, _ := addr.IAFromString("2-ff00:0:110")
// 	IA3, _ := addr.IAFromString("3-ff00:0:110")
// 	IA4, _ := addr.IAFromString("4-ff00:0:110")

// 	rul1 := InternalClassRule{Name: "Hello Test", SourceAs: matchRule{IA: IA1}}
// 	rul2 := InternalClassRule{Name: "Hello World", SourceAs: matchRule{IA: IA2}}
// 	rul3 := InternalClassRule{Name: "Hello SCION", SourceAs: matchRule{IA: IA3}}
// 	rul4 := InternalClassRule{Name: "Hello Internet", SourceAs: matchRule{IA: IA4}}

// 	m[IA1] = &rul1
// 	m[IA2] = &rul2
// 	m[IA3] = &rul3
// 	m[IA4] = &rul4

// 	search, _ := addr.IAFromString("3-ff00:0:110")

// 	rule, found := m[search]
// 	fmt.Println("We have found", found, rule)

// 	search, _ = addr.IAFromString("5-ff00:0:110")

// 	rule, found = m[search]
// 	fmt.Println("We have found", found, rule)

// 	t.Errorf("See logs")

// }

// func TestLoadingToMaps(t *testing.T) {
// 	r, _ := setupTestRouter(t)

// 	r.initQueueing("sample-config.yaml")

// 	r.config.SourceRules, r.config.DestinationRules = rulesToMap(r.config.Rules)

// 	search, _ := addr.IAFromString("1-ff00:0:110")
// 	rule, found := r.config.SourceRules[search]
// 	fmt.Println("We have found", found, rule)

// 	if !found {
// 		t.Errorf("See logs")
// 	}

// 	search, _ = addr.IAFromString("5-ff00:0:110")
// 	rule, found = r.config.SourceRules[search]
// 	fmt.Println("We have found", found, rule)

// 	if found {
// 		t.Errorf("See logs")
// 	}

// }

// func TestMatchingRules(t *testing.T) {

// 	r, _ := setupTestRouter(t)

// 	r.initQueueing("sample-config.yaml")

// 	srcAddr, _ := addr.IAFromString("1-ff00:0:110")
// 	dstAddr, _ := addr.IAFromString("1-ff00:0:111")

// 	queues1 := r.config.SourceRules[srcAddr]
// 	queues2 := r.config.DestinationRules[dstAddr]

// 	for _, rul1 := range queues1 {
// 		for _, rul2 := range queues2 {
// 			if rul1 == rul2 {
// 				if rul1.QueueNumber != 2 {
// 					t.Errorf("See logs")
// 				}
// 			}
// 		}
// 	}
// }
