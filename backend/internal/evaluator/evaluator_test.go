package evaluator

import (
	"errors"
	"testing"
	"uuid"
)

func TestEvaluate(t *testing.T) {
	mustNewRule := func(id uuid.UUID, serve float64, clauses ...Clause) Rule {
		r, err := NewRule(id, clauses, serve)
		if err != nil {
			t.Fatalf("NewRule(): %v", err)
		}
		return r
	}
	mustNewFlagConfig := func(enabled bool, allowlist []string, fall float64, rules ...Rule) FlagConfig {
		users := make(map[string]struct{}, len(allowlist))
		for _, key := range allowlist {
			users[key] = struct{}{}
		}
		cfg, err := NewFlagConfig(enabled, "s4lt", "new-checkout", users, rules, fall)
		if err != nil {
			t.Fatalf("NewFlagConfig(): %v", err)
		}
		return cfg
	}

	proID := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-000000000001")
	countryID := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-000000000002")
	offID := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-000000000003")
	pro := mustNewRule(proID, 1, mustNewClause(t, "isPro", eqOp, BooleanClause, true))
	country := mustNewRule(countryID, 0.5, mustNewClause(t, "country", eqOp, StringClause, "BJ"))
	off := mustNewRule(offID, 0, mustNewClause(t, "country", eqOp, StringClause, "FR"))

	alice := "user-alice"
	bob := "user-bob"

	tests := map[string]struct {
		config     FlagConfig
		attributes map[string][]any
		wantServe  float64
		wantReason string
		errStr     string
	}{
		// The master switch wins over everything, including bad input.
		"disabled flag serves 0": {
			config:     mustNewFlagConfig(false, []string{alice}, 1, pro),
			attributes: map[string][]any{"key": {alice}, "isPro": {true}},
			wantServe:  0,
			wantReason: MasterSwitchReason,
		},
		"disabled flag ignores an invalid key": {
			config:     mustNewFlagConfig(false, nil, 1, pro),
			attributes: map[string][]any{"key": {42.0}},
			wantServe:  0,
			wantReason: MasterSwitchReason,
		},
		"disabled flag ignores erroring rules": {
			config:     mustNewFlagConfig(false, nil, 1, pro),
			attributes: map[string][]any{"isPro": {"true"}},
			wantServe:  0,
			wantReason: MasterSwitchReason,
		},

		"allowlisted key serves 1": {
			config:     mustNewFlagConfig(true, []string{bob, alice}, 0, off),
			attributes: map[string][]any{"key": {alice}},
			wantServe:  1,
			wantReason: AllowlistMatchReason,
		},
		// The allowlist bypasses the rules, even one that would serve 0.
		"allowlist checked before rules": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, off),
			attributes: map[string][]any{"key": {alice}, "country": {"FR"}},
			wantServe:  1,
			wantReason: AllowlistMatchReason,
		},
		"allowlist checked before erroring rules": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, pro),
			attributes: map[string][]any{"key": {alice}, "isPro": {"true"}},
			wantServe:  1,
			wantReason: AllowlistMatchReason,
		},
		"key not in allowlist goes on to rules": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, country),
			attributes: map[string][]any{"key": {bob}, "country": {"BJ"}},
			wantServe:  0.5,
			wantReason: RuleMatchReason + countryID.String(),
		},
		"missing key goes on to rules": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, country),
			attributes: map[string][]any{"country": {"BJ"}},
			wantServe:  0.5,
			wantReason: RuleMatchReason + countryID.String(),
		},
		"key with multiple values rejected": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, country),
			attributes: map[string][]any{"key": {alice, bob}},
			errStr:     "the key attribute must have exactly one value, got: 2",
		},
		"non-string key rejected": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, country),
			attributes: map[string][]any{"key": {true}},
			errStr:     "the key attribute must be a string",
		},
		"key with no value rejected": {
			config:     mustNewFlagConfig(true, []string{alice}, 0, country),
			attributes: map[string][]any{"key": {}},
			errStr:     "the key attribute must have exactly one value, got: 0",
		},
		// The key is validated even when there is no allowlist to check it against.
		"invalid key rejected without an allowlist": {
			config:     mustNewFlagConfig(true, nil, 0, country),
			attributes: map[string][]any{"key": {42.0}, "country": {"BJ"}},
			errStr:     "the key attribute must be a string",
		},

		"matching rule serves its value": {
			config:     mustNewFlagConfig(true, nil, 0, pro),
			attributes: map[string][]any{"isPro": {true}},
			wantServe:  1,
			wantReason: RuleMatchReason + proID.String(),
		},
		// A rule serving 0 is still a match, not a fallthrough.
		"matching rule serving 0": {
			config:     mustNewFlagConfig(true, nil, 1, off),
			attributes: map[string][]any{"country": {"FR"}},
			wantServe:  0,
			wantReason: RuleMatchReason + offID.String(),
		},
		"first matching rule wins": {
			config:     mustNewFlagConfig(true, nil, 0, pro, country),
			attributes: map[string][]any{"isPro": {true}, "country": {"BJ"}},
			wantServe:  1,
			wantReason: RuleMatchReason + proID.String(),
		},
		"later rule matches when earlier ones don't": {
			config:     mustNewFlagConfig(true, nil, 0, pro, off, country),
			attributes: map[string][]any{"isPro": {false}, "country": {"BJ"}},
			wantServe:  0.5,
			wantReason: RuleMatchReason + countryID.String(),
		},
		// Rules after the first match are not evaluated, so their errors
		// never surface.
		"match stops before a later erroring rule": {
			config:     mustNewFlagConfig(true, nil, 0, country, pro),
			attributes: map[string][]any{"country": {"BJ"}, "isPro": {"true"}},
			wantServe:  0.5,
			wantReason: RuleMatchReason + countryID.String(),
		},

		"no rule matches serves fall": {
			config:     mustNewFlagConfig(true, []string{alice}, 0.2, pro, country),
			attributes: map[string][]any{"key": {bob}, "isPro": {false}, "country": {"US"}},
			wantServe:  0.2,
			wantReason: FallthroughReason,
		},
		"no rules serves fall": {
			config:     mustNewFlagConfig(true, nil, 0.7),
			attributes: map[string][]any{"country": {"BJ"}},
			wantServe:  0.7,
			wantReason: FallthroughReason,
		},
		"no attributes serves fall": {
			config:     mustNewFlagConfig(true, []string{alice}, 0.3, pro),
			attributes: map[string][]any{},
			wantServe:  0.3,
			wantReason: FallthroughReason,
		},

		// On a rule error the flag serves 0, not its fallthrough value.
		"rule error is wrapped with its index and serves 0": {
			config:     mustNewFlagConfig(true, nil, 0.4, country, pro),
			attributes: map[string][]any{"country": {"FR"}, "isPro": {"true"}},
			wantServe:  0,
			errStr:     `rule at index 1 errors with: clause at index 0 errors: invalid value for attribute "isPro": invalid value true`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			serve, reason, err := Evaluate(tc.config, tc.attributes)
			if tc.errStr != "" {
				if err == nil || err.Error() != tc.errStr {
					t.Fatalf("Evaluate() error = %v, want %q", err, tc.errStr)
				}
				if serve != tc.wantServe || reason != "" {
					t.Fatalf("Evaluate() = (%v, %q) on error, want (%v, \"\")", serve, reason, tc.wantServe)
				}
				return
			}
			if err != nil {
				t.Fatalf("Evaluate() unexpected error: %v", err)
			}
			if serve != tc.wantServe || reason != tc.wantReason {
				t.Fatalf("Evaluate() = (%v, %q), want (%v, %q)", serve, reason, tc.wantServe, tc.wantReason)
			}
		})
	}
}

func TestEvaluateUnwrapsRuleError(t *testing.T) {
	errBoom := errors.New("boom")
	rule, err := NewRule(uuid.New(), []Clause{stubClause{err: errBoom}}, 1)
	if err != nil {
		t.Fatalf("NewRule(): %v", err)
	}
	cfg, err := NewFlagConfig(true, "s4lt", "new-checkout", nil, []Rule{rule}, 0)
	if err != nil {
		t.Fatalf("NewFlagConfig(): %v", err)
	}

	_, _, err = Evaluate(cfg, nil)
	if !errors.Is(err, errBoom) {
		t.Fatalf("Evaluate() error = %v, want it to wrap %v", err, errBoom)
	}
}
