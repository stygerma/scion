package conf

import (
	"io/ioutil"

	"github.com/scionproto/scion/go/lib/log"
	"gopkg.in/yaml.v2"
)

type PoliceAction uint8

// Actions to execute on packets in queues.
const (
	PASS PoliceAction = iota
	NOTIFY
	DROP
	DROPNOTIFY
)

type CongestionWarning struct {
	Approach           int `yaml:"approach"`
	InformationContent int `yaml:"informationContent"`
}

type ActionProfile struct {
	FillLevel int          `yaml:"fill-level"`
	Prob      int          `yaml:"prob"`
	Action    PoliceAction `yaml:"action"`
}

type ExternalPacketQueue struct {
	Name              string            `yaml:"name"`
	ID                int               `yaml:"id"`
	MinBandwidth      int               `yaml:"CIR"`
	MaxBandWidth      int               `yaml:"PIR"`
	PoliceRate        string            `yaml:"policeRate"`
	MaxLength         int               `yaml:"maxLength"`
	Priority          int               `yaml:"priority"`
	CongestionWarning CongestionWarning `yaml:"congestionWarning"`
	Profile           []ActionProfile   `yaml:"profile"`
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
	var yamlFile []byte
	var err error

	// yamlFile, err = ioutil.ReadFile(configFileLocation)
	yamlFile, err = ioutil.ReadFile(path)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		yamlFile, err = ioutil.ReadFile("/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/qos/sample-config.yaml")
	}

	if err != nil {
		log.Error("Loading the config file has failed", "error", err)
		return ExternalConfig{}, err
	}
	err = yaml.Unmarshal(yamlFile, &ec)
	if err != nil {
		log.Error("Loading the config file has failed", "error", err)
		return ExternalConfig{}, err
	}
	return ec, nil
}
