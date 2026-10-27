package evaluator

import (
	"fmt"
	"maps"
	"slices"
	"uuid"
)

// FlagConfig represents the flag setting
type FlagConfig struct {
	// enabled is the master switch. When true we procces with the
	// allowlist, rules checks
	enabled bool
	// allowlist is a slice of user IDs authorize to get the feature
	// bypassing the rules
	allowlist map[uuid.UUID]struct{}
	// rules represents the checks deciding the whether or not the current
	// user can get the feature
	rules []Rule
	// fall is the fallthrough result when all the checks fail
	// to answer. It's a rollout percentage of inclusive range [0, 1]
	fall float64
}

// NewFlagConfig ensure the creation of a valid FlagConfig
func NewFlagConfig(enabled bool, allowlist map[uuid.UUID]struct{}, rules []Rule, fall float64) (FlagConfig, error) {
	rulesCopy := slices.Clone(rules)
	allowlistCopy := maps.Clone(allowlist)

	if !(0 <= fall && fall <= 1) {
		return FlagConfig{}, fmt.Errorf("invalid fallthrough value: %f", fall)
	}
	for i, rule := range rulesCopy {
		if len(rule.clauses) == 0 {
			return FlagConfig{}, fmt.Errorf("invalid rule at index %d", i)
		}
	}
	for key := range allowlistCopy {
		if key == uuid.Nil() {
			return FlagConfig{}, fmt.Errorf("invalid allowlist element %q", key)
		}
	}

	return FlagConfig{
		enabled:   enabled,
		allowlist: allowlistCopy,
		rules:     rulesCopy,
		fall:      fall,
	}, nil
}
