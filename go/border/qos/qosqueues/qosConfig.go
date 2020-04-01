package qosqueues

import (
	"github.com/scionproto/scion/go/lib/addr"
)

// InternalRouterConfig is what I am loading from the config file
// type InternalRouterConfig struct {
// 	Queues           []PacketQueueInterface
// 	Rules            []InternalClassRule
// 	SourceRules      map[addr.IA][]*InternalClassRule
// 	DestinationRules map[addr.IA][]*InternalClassRule
// }

type InternalRouterConfig struct {
	Queues []PacketQueueInterface
	Rules  MapRules
}

type MapRules struct {
	RulesList                 []InternalClassRule
	SourceRules               map[addr.IA][]*InternalClassRule
	DestinationRules          map[addr.IA][]*InternalClassRule
	SourceAnyDestinationRules map[addr.IA][]*InternalClassRule
	DestinationAnySourceRules map[addr.IA][]*InternalClassRule
	ASOnlySourceRules         map[addr.AS][]*InternalClassRule
	ASOnlyDestRules           map[addr.AS][]*InternalClassRule
	ISDOnlySourceRules        map[addr.ISD][]*InternalClassRule
	ISDOnlyDestRules          map[addr.ISD][]*InternalClassRule
}
