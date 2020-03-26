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
