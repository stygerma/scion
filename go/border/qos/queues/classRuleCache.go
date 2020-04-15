package queues

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

type cacheEntry struct {
	srcAddress addr.IA
	dstAddress addr.IA
	l4type     common.L4ProtocolType
}

type ClassRuleCacheInterface interface {
	Init(maxEntries int)
	Get(entry cacheEntry) *InternalClassRule
	Put(entry cacheEntry, rule *InternalClassRule)
}

type ClassRuleCache struct {
	cacheMap map[cacheEntry]*InternalClassRule
}

func (crCache *ClassRuleCache) Init(maxEntries int) {
	crCache.cacheMap = make(map[cacheEntry]*InternalClassRule, maxEntries)
}

func (crCache *ClassRuleCache) Get(entry cacheEntry) *InternalClassRule {

	return crCache.cacheMap[entry]
}

func (crCache *ClassRuleCache) Put(entry cacheEntry, rule *InternalClassRule) {

	crCache.cacheMap[entry] = rule
}
