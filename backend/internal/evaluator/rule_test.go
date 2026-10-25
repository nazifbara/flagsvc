package evaluator

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewRule(t *testing.T) {
	country := mustNewClause(t, "country", eqOp, StringClause, "BJ")
	age := mustNewClause(t, "age", gteOp, NumberClause, 18.0)
	isPro := mustNewClause(t, "isPro", eqOp, BooleanClause, true)

	tests := map[string]struct {
		clauses []Clause
		serve   float64
		want    Rule
		errStr  string
	}{
		"nil clauses rejected": {
			clauses: nil,
			serve:   1,
			errStr:  "can't build a rule with no clause",
		},
		"empty clauses rejected": {
			clauses: []Clause{},
			serve:   1,
			errStr:  "can't build a rule with no clause",
		},
		"nil entry rejected": {
			clauses: []Clause{country, nil, age},
			serve:   1,
			errStr:  "nil clause at index 1",
		},
		"nil entry checked before serve": {
			clauses: []Clause{nil},
			serve:   2,
			errStr:  "nil clause at index 0",
		},
		"negative serve rejected": {
			clauses: []Clause{country},
			serve:   -0.1,
			errStr:  "invalid serve value: -0.100000",
		},
		"serve above 1 rejected": {
			clauses: []Clause{country},
			serve:   1.1,
			errStr:  "invalid serve value: 1.100000",
		},
		// NaN fails every comparison, so a plain range check lets it through.
		"NaN serve rejected": {
			clauses: []Clause{country},
			serve:   math.NaN(),
			errStr:  "invalid serve value: NaN",
		},
		"empty clauses checked before serve": {
			clauses: []Clause{},
			serve:   2,
			errStr:  "can't build a rule with no clause",
		},

		// Both bounds are inclusive.
		"serve 0 accepted": {
			clauses: []Clause{age},
			serve:   0,
			want:    Rule{clauses: []Clause{age}, serve: 0},
		},
		"serve 1 accepted": {
			clauses: []Clause{age},
			serve:   1,
			want:    Rule{clauses: []Clause{age}, serve: 1},
		},
		"partial serve accepted": {
			clauses: []Clause{country},
			serve:   0.25,
			want:    Rule{clauses: []Clause{country}, serve: 0.25},
		},
		"multiple clauses keep their order": {
			clauses: []Clause{isPro, country, age},
			serve:   0.5,
			want:    Rule{clauses: []Clause{isPro, country, age}, serve: 0.5},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewRule(tc.clauses, tc.serve)
			if tc.errStr != "" {
				if err == nil || err.Error() != tc.errStr {
					t.Fatalf("NewRule() error = %v, want %q", err, tc.errStr)
				}
				opt := cmp.AllowUnexported(Rule{})
				if diff := cmp.Diff(Rule{}, got, opt); diff != "" {
					t.Fatalf("NewRule() = %#v on error, want zero Rule", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewRule() unexpected error: %v", err)
			}

			opt := cmp.AllowUnexported(Rule{}, stringClause{}, booleanClause{}, numberClause{})
			if diff := cmp.Diff(tc.want, got, opt); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestNewRuleCopiesClauses(t *testing.T) {
	country := mustNewClause(t, "country", eqOp, StringClause, "BJ")
	age := mustNewClause(t, "age", gteOp, NumberClause, 18.0)

	clauses := []Clause{country}
	rule, err := NewRule(clauses, 1)
	if err != nil {
		t.Fatalf("NewRule() unexpected error: %v", err)
	}
	clauses[0] = age

	opt := cmp.AllowUnexported(stringClause{}, numberClause{})
	if diff := cmp.Diff([]Clause{country}, rule.clauses, opt); diff != "" {
		t.Fatalf("rule changed after caller modified its slice:\n%s", diff)
	}
}
