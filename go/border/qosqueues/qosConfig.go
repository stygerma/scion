package qosqueues

import (
	"github.com/scionproto/scion/go/lib/addr"
)

type Surplus struct {
	Surplus  int
	Payments []int
}

// InternalRouterConfig is what I am loading from the config file
type InternalRouterConfig struct {
	Queues           []PacketQueueInterface
	Rules            []InternalClassRule
	SourceRules      map[addr.IA][]*InternalClassRule
	DestinationRules map[addr.IA][]*InternalClassRule
}

// RouterConfig is what I am loading from the config file
type RouterConfig struct {
	Queues []PacketQueue `yaml:"Queues"`
	Rules  []classRule   `yaml:"Rules"`
}
