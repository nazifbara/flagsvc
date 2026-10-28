package evaluator

import (
	"errors"
	"fmt"
)

const (
	MasterSwitchReason   = "master_switch"
	AllowlistMatchReason = "allowlist_match"
	RuleMatchReason      = "rule_match="
	FallthroughReason    = "fallthrough"
)

func Evaluate(flagConfig FlagConfig, attributes map[string][]any) (serve float32, reason string, err error) {
	if !flagConfig.enabled {
		return 0, MasterSwitchReason, nil
	}
	key, ok := attributes["key"]
	if ok {
		if len(key) != 1 {
			return 0, "", fmt.Errorf("the key attribute must have exactly one value, got: %d", len(key))
		}
		keyString, ok := key[0].(string)
		if !ok {
			return 0, "", errors.New("the key attribute must be a string")
		}
		if _, ok = flagConfig.allowlist[keyString]; ok {
			return 1, AllowlistMatchReason, nil
		}
	} else {
		return 0, "", errors.New("the key string attributes is required")
	}

	for i, rule := range flagConfig.rules {
		serve, matched, err := rule.Match(attributes)
		if err != nil {
			return 0, "", fmt.Errorf("rule at index %d errors with: %w", i, err)
		}
		if matched {
			return serve, RuleMatchReason + rule.id.String(), nil
		}
	}

	return flagConfig.fall, FallthroughReason, nil
}
