package qosconf

import (
	"io/ioutil"

	"github.com/scionproto/scion/go/lib/log"
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

type SchedulerConfig struct {
	Latency   int    `yaml:"Latency"`
	Bandwidth string `yaml:"Bandwidth"`
}

const configFileLocation = "/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/qos/sample-config.yaml"

// ExternalConfig is what I am loading from the config file
type ExternalConfig struct {
	SchedulerConfig SchedulerConfig       `yaml:"Scheduler"`
	ExternalQueues  []ExternalPacketQueue `yaml:"Queues"`
	ExternalRules   []ExternalClassRule   `yaml:"Rules"`
}

func LoadConfig(path string) (ExternalConfig, error) {

	var ec ExternalConfig
	var yamlFile []byte
	var err error

	yamlFile, err = ioutil.ReadFile(configFileLocation)
	// yamlFile, err = ioutil.ReadFile(path)

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

	log.Info("Config File is", "ec", ec)

	return ec, nil

}
