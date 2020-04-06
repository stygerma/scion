package qosconf

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type PoliceAction uint8

const (
	// PASS Pass the packet
	PASS PoliceAction = iota
	// NOTIFY Notify the sending host of the packet
	NOTIFY
	// DROP Drop the packet
	DROP
	// DROPNOTIFY Drop and then notify someone
	DROPNOTIFY
)

type ActionProfile struct {
	FillLevel int          `yaml:"fill-level"`
	Prob      int          `yaml:"prob"`
	Action    PoliceAction `yaml:"action"`
}

type ExternalPacketQueue struct {
	Name         string          `yaml:"name"`
	ID           int             `yaml:"id"`
	MinBandwidth int             `yaml:"CIR"`
	MaxBandWidth int             `yaml:"PIR"`
	PoliceRate   string          `yaml:"policeRate"`
	MaxLength    int             `yaml:"maxLength"`
	Priority     int             `yaml:"priority"`
	Profile      []ActionProfile `yaml:"profile"`
}

// ExternalClassRule contains a rule for matching packets
type ExternalClassRule struct {
	// This is currently means the ID of the sending border router
	Name                 string `yaml:"name"`
	Priority             int    `yaml:"priority"`
	SourceAs             string `yaml:"sourceAs"`
	SourceMatchMode      int    `yaml:"sourceMatchMode"`
	DestinationAs        string `yaml:"destinationAs"`
	DestinationMatchMode int    `yaml:"destinationMatchMode"`
	L4Type               []int  `yaml:"L4Type"`
	QueueNumber          int    `yaml:"queueNumber"`
}

// ExternalConfig is what I am loading from the config file
type ExternalConfig struct {
	ExternalQueues []ExternalPacketQueue `yaml:"Queues"`
	ExternalRules  []ExternalClassRule   `yaml:"Rules"`
}

func LoadConfig(path string) (ExternalConfig, error) {
	var ec ExternalConfig

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return ExternalConfig{}, err
	}
	err = yaml.Unmarshal(yamlFile, &ec)
	if err != nil {
		return ExternalConfig{}, err
	}
	return ec, nil
}
