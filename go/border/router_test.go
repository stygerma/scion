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

	fmt.Println("The Queue is: ", r.config.Queues[0])
	fmt.Println("The Rule is: ", r.config.Rules[0])
	// t.Errorf("Output: %v", r.config)

}

func TestLoadSampleConfigQueues(t *testing.T) {
	r, _ := setupTestRouter(t)

	r.loadConfigFile("sample-config.yaml")

	fmt.Println("The Queue is: ", r.config.Queues[0])
	fmt.Println("The Rule is: ", r.config.Rules[0])
	fmt.Println("We have this number of queues: ", len(r.config.Queues))
	t.Errorf("Output: %v", r.config)

	if r.config.Queues[0].ID != 0 {
		t.Errorf("Incorrect Queue ID")
	}
	if r.config.Queues[1].ID != 1 {
		t.Errorf("Incorrect Queue ID")
	}

	if r.config.Queues[0].Name != "General Queue" {
		t.Errorf("Incorrect Queue Name is %v but should be %v", r.config.Queues[0].Name, "General Queue")
	}
	if r.config.Queues[1].Name != "Speedy Queue" {
		t.Errorf("Incorrect Queue Name is %v but should be %v", r.config.Queues[0].Name, "Speedy Queue")
	}
}
