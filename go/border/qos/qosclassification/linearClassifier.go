package qosLinearClassifier

import (
	"strings"

	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

type LinearClassifier struct{}

var _ ClassifierInterface = (*LinearClassifier)(nil)

func getMatchFromRule(cr qosconf.ExternalClassRule, matchModeField int, matchRuleField string) (matchRule, error) {
	switch matchMode(matchModeField) {
	case EXACT, ASONLY, ISDONLY, ANY:
		IA, err := addr.IAFromString(matchRuleField)
		if err != nil {
			return matchRule{}, err
		}
		m := matchRule{IA: IA, lowLim: addr.IA{}, upLim: addr.IA{}, matchMode: matchMode(matchModeField)}
		return m, nil
	case RANGE:
		if matchMode(matchModeField) == RANGE {
			parts := strings.Split(matchRuleField, "||")
			if len(parts) != 2 {
				return matchRule{}, common.NewBasicError("Invalid Class", nil, "raw", matchModeField)
			}
			lowLim, err := addr.IAFromString(parts[1])
			if err != nil {
				return matchRule{}, err
			}
			upLim, err := addr.IAFromString(parts[1])
			if err != nil {
				return matchRule{}, err
			}
			m := matchRule{IA: addr.IA{}, lowLim: lowLim, upLim: upLim, matchMode: matchMode(matchModeField)}
			return m, nil
		}
	}

	return matchRule{}, common.NewBasicError("Invalid matchMode declared", nil, "matchMode", matchModeField)
}

func getQueueNumberIterativeForInternal(config *InternalRouterConfig, rp *rpkt.RtrPkt) int {

	queueNo := 0
	matches := make([]InternalClassRule, 0)

	for _, cr := range config.Rules {

		if cr.matchInternalRule(rp) {
			matches = append(matches, cr)
		}
	}

	max := -1
	for _, rul1 := range matches {
		if rul1.Priority > max {
			queueNo = rul1.QueueNumber
			max = rul1.Priority
		}
	}

	return queueNo
}

func getQueueNumberIterativeFor(legacyConfig *qosconf.ExternalConfig, rp *rpkt.RtrPkt) int {
	queueNo := 0

	matches := make([]qosconf.ExternalClassRule, 0)

	for _, cr := range legacyConfig.ExternalRules {
		if matchRuleFromConfig(&cr, rp) {
			matches = append(matches, cr)
		}
	}

	max := -1
	for _, rul1 := range matches {
		if rul1.Priority > max {
			queueNo = rul1.QueueNumber
			max = rul1.Priority
		}
	}

	return queueNo
}
