package evaluator

import (
	"math"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

func TestNewFlagConfig(t *testing.T) {
	mustNewRule := func(serve float64, clauses ...Clause) Rule {
		r, err := NewRule(uuid.New(), clauses, serve)
		if err != nil {
			t.Fatalf("NewRule(): %v", err)
		}
		return r
	}
	country := mustNewRule(1, mustNewClause(t, "country", eqOp, StringClause, "BJ"))
	adults := mustNewRule(0.5, mustNewClause(t, "age", gteOp, NumberClause, 18.0))
	alice := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-2a3b4c5d6e7f")
	bob := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-2a3b4c5d6e80")
	users := func(ids ...uuid.UUID) map[uuid.UUID]struct{} {
		m := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			m[id] = struct{}{}
		}
		return m
	}

	tests := map[string]struct {
		enabled   bool
		allowlist map[uuid.UUID]struct{}
		rules     []Rule
		fall      float64
		want      FlagConfig
		errStr    string
	}{
		"negative fall rejected": {
			enabled: true,
			rules:   []Rule{country},
			fall:    -0.1,
			errStr:  "invalid fallthrough value: -0.100000",
		},
		"fall above 1 rejected": {
			enabled: true,
			rules:   []Rule{country},
			fall:    1.1,
			errStr:  "invalid fallthrough value: 1.100000",
		},
		"NaN fall rejected": {
			enabled: true,
			rules:   []Rule{country},
			fall:    math.NaN(),
			errStr:  "invalid fallthrough value: NaN",
		},
		// A Rule with no clauses matches every user, so it would shadow the
		// rules after it.
		"zero rule rejected": {
			enabled: true,
			rules:   []Rule{country, {}, adults},
			fall:    0,
			errStr:  "invalid rule at index 1",
		},
		"nil allowlist ID rejected": {
			enabled:   true,
			allowlist: users(alice, uuid.Nil()),
			fall:      0,
			errStr:    `invalid allowlist element "00000000-0000-0000-0000-000000000000"`,
		},
		"fall checked before rules": {
			enabled: true,
			rules:   []Rule{{}},
			fall:    2,
			errStr:  "invalid fallthrough value: 2.000000",
		},

		// Both bounds of fall are inclusive.
		"fall 0 accepted": {
			enabled: true,
			rules:   []Rule{country},
			fall:    0,
			want:    FlagConfig{enabled: true, rules: []Rule{country}, fall: 0},
		},
		"fall 1 accepted": {
			enabled: true,
			rules:   []Rule{country},
			fall:    1,
			want:    FlagConfig{enabled: true, rules: []Rule{country}, fall: 1},
		},
		"partial fall accepted": {
			enabled:   true,
			allowlist: users(alice),
			rules:     []Rule{country},
			fall:      0.1,
			want:      FlagConfig{enabled: true, allowlist: users(alice), rules: []Rule{country}, fall: 0.1},
		},
		"disabled flag keeps its settings": {
			enabled:   false,
			allowlist: users(alice),
			rules:     []Rule{country},
			fall:      0.1,
			want:      FlagConfig{enabled: false, allowlist: users(alice), rules: []Rule{country}, fall: 0.1},
		},
		// Rules are evaluated in order, so the order given must be kept.
		"rules keep their order": {
			enabled:   true,
			allowlist: users(alice, bob),
			rules:     []Rule{adults, country},
			fall:      0.5,
			want:      FlagConfig{enabled: true, allowlist: users(alice, bob), rules: []Rule{adults, country}, fall: 0.5},
		},
		"no allowlist or rules, only fall": {
			enabled: true,
			fall:    1,
			want:    FlagConfig{enabled: true, fall: 1},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewFlagConfig(tc.enabled, tc.allowlist, tc.rules, tc.fall)
			if tc.errStr != "" {
				if err == nil || err.Error() != tc.errStr {
					t.Fatalf("NewFlagConfig() error = %v, want %q", err, tc.errStr)
				}
				opt := cmp.AllowUnexported(FlagConfig{})
				if diff := cmp.Diff(FlagConfig{}, got, opt); diff != "" {
					t.Fatalf("NewFlagConfig() = %#v on error, want zero FlagConfig", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFlagConfig() unexpected error: %v", err)
			}

			opt := cmp.AllowUnexported(FlagConfig{}, Rule{}, stringClause{}, numberClause{})
			if diff := cmp.Diff(tc.want, got, opt); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestNewFlagConfigCopiesInputs(t *testing.T) {
	country, err := NewRule(uuid.New(), []Clause{mustNewClause(t, "country", eqOp, StringClause, "BJ")}, 1)
	if err != nil {
		t.Fatalf("NewRule(): %v", err)
	}
	adults, err := NewRule(uuid.New(), []Clause{mustNewClause(t, "age", gteOp, NumberClause, 18.0)}, 0.5)
	if err != nil {
		t.Fatalf("NewRule(): %v", err)
	}
	alice := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-2a3b4c5d6e7f")
	bob := uuid.MustParse("0192e4a0-7b3c-7d4e-9f10-2a3b4c5d6e80")

	allowlist := map[uuid.UUID]struct{}{alice: {}}
	rules := []Rule{country}
	cfg, err := NewFlagConfig(true, allowlist, rules, 0)
	if err != nil {
		t.Fatalf("NewFlagConfig() unexpected error: %v", err)
	}
	allowlist[bob] = struct{}{}
	rules[0] = adults

	if diff := cmp.Diff(map[uuid.UUID]struct{}{alice: {}}, cfg.allowlist); diff != "" {
		t.Errorf("allowlist changed after caller modified its map:\n%s", diff)
	}
	opt := cmp.AllowUnexported(Rule{}, stringClause{}, numberClause{})
	if diff := cmp.Diff([]Rule{country}, cfg.rules, opt); diff != "" {
		t.Errorf("rules changed after caller modified its slice:\n%s", diff)
	}
}
