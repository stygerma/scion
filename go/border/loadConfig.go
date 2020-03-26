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
type configFileRouterConfig struct {
	Queues []configFilePacketQueue `yaml:"Queues"`
	Rules  []configFileClassRule   `yaml:"Rules"`
}

type configFileActionProfile struct {
	FillLevel int                    `yaml:"fill-level"`
	Prob      int                    `yaml:"prob"`
	Action    qosqueues.PoliceAction `yaml:"action"`
}

type configFilePacketQueue struct {
	Name         string                    `yaml:"name"`
	ID           int                       `yaml:"id"`
	MinBandwidth int                       `yaml:"CIR"`
	MaxBandWidth int                       `yaml:"PIR"`
	PoliceRate   int                       `yaml:"policeRate"`
	MaxLength    int                       `yaml:"maxLength"`
	Priority     int                       `yaml:"priority"`
	Profile      []configFileActionProfile `yaml:"profile"`
}

type configFileClassRule struct {
	// This is currently means the ID of the sending border router
	Name                 string `yaml:"name"`
	SourceAs             string `yaml:"sourceAs"`
	SourceMatchMode      int    `yaml:"sourceMatchMode"`
	NextHopAs            string `yaml:"nextHopAs"`
	NextHopMatchMode     int    `yaml:"nextHopMatchMode"`
	DestinationAs        string `yaml:"destinationAs"`
	DestinationMatchMode int    `yaml:"destinationMatchMode"`
	L4Type               []int  `yaml:"L4Type"`
	QueueNumber          int    `yaml:"queueNumber"`
}

func (r *Router) loadConfigFile(path string) error {

	var internalRules []classRule
	var internalQueues []qosqueues.PacketQueueInterface

	var rc configFileRouterConfig

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
		queueToUse.InitQueue(convertConfigFileQueueToQueue(que), muta, mutb)
		internalQueues = append(internalQueues, queueToUse)
	}

	r.config = routerConfig{Queues: internalQueues, Rules: internalRules}

	return nil
}

func convertConfigFileQueueToQueue(cfQueue configFilePacketQueue) qosqueues.PacketQueue {

	var ap []qosqueues.ActionProfile

	for _, prof := range cfQueue.Profile {
		intProf := qosqueues.ActionProfile{
			FillLevel: prof.FillLevel,
			Prob:      prof.Prob,
			Action:    prof.Action,
		}
		ap = append(ap, intProf)
	}

	que := qosqueues.PacketQueue{
		Name:         cfQueue.Name,
		ID:           cfQueue.ID,
		MinBandwidth: cfQueue.MinBandwidth,
		MaxBandWidth: cfQueue.MaxBandWidth,
		PoliceRate:   cfQueue.PoliceRate,
		MaxLength:    cfQueue.MaxLength,
		Priority:     cfQueue.Priority,
		Profile:      ap,
	}

	return que
}
