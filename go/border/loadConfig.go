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
	"sync"

	"github.com/scionproto/scion/go/border/qosqueues"

	"github.com/scionproto/scion/go/lib/log"
	"gopkg.in/yaml.v2"
)

// RouterConfig is what I am loading from the config file
type RouterConfig struct {
	Queues []qosqueues.PacketQueue `yaml:"Queues"`
	Rules  []classRule             `yaml:"Rules"`
}

func (r *Router) loadConfigFile(path string) error {

	var internalRules []internalClassRule
	var internalQueues []qosqueues.PacketQueueInterface

	var rc RouterConfig

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
		intRule, err := convClassRuleToInternal(rule)
		if err != nil {
			log.Error("Error reading config file", "error", err)
		}
		internalRules = append(internalRules, intRule)
	}

	for _, que := range rc.Queues {
		muta := &sync.Mutex{}
		mutb := &sync.Mutex{}
		queueToUse.InitQueue(que, muta, mutb)
		internalQueues = append(internalQueues, queueToUse)
	}

	r.legacyConfig = rc
	r.config = InternalRouterConfig{Queues: internalQueues, Rules: internalRules}

	return nil
}
