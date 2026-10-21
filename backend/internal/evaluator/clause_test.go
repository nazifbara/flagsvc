package evaluator

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type newClauseTest struct {
	attr   string
	op     string
	kind   ClauseKind
	values []any
	want   Clause
	errStr string
}

func TestNewClause(t *testing.T) {
	tests := map[string]newClauseTest{
		"unknown operator": {
			attr:   "abc",
			op:     "equal",
			kind:   BooleanClause,
			values: []any{true},
			errStr: "unknown operator: equal",
		},
		"eq operator not supported on multi-value clause": {
			attr:   "abc",
			op:     eqOp,
			kind:   NumberClause,
			values: []any{1.0, 2.0},
			errStr: "eq operator on multi-value clause",
		},
		"neq operator not supported on multi-value clause": {
			attr:   "abc",
			op:     neqOp,
			kind:   StringClause,
			values: []any{"a", "b"},
			errStr: "neq operator on multi-value clause",
		},
		"il operator accepts multiple values": {
			attr:   "country",
			op:     ilOp,
			kind:   StringClause,
			values: []any{"BJ", "FR"},
			want:   Clause{attribute: "country", operator: ilOp, kind: StringClause, values: []any{"BJ", "FR"}},
		},
		"cnt operator accepts multiple values": {
			attr:   "country",
			op:     cntOp,
			kind:   StringClause,
			values: []any{"BJ", "FR"},
			want:   Clause{attribute: "country", operator: cntOp, kind: StringClause, values: []any{"BJ", "FR"}},
		},
		"StringClause: invalid value type": {
			attr:   "abc",
			op:     ilOp,
			kind:   StringClause,
			values: []any{"1", 2, 3},
			errStr: "invalid value for StringClause",
		},
		"StringClause: invalid value checked before unsupported operator": {
			attr:   "role",
			op:     gtOp,
			kind:   StringClause,
			values: []any{1.0},
			errStr: "invalid value for StringClause",
		},
		"BooleanClause: invalid value type": {
			attr:   "isPro",
			op:     eqOp,
			kind:   BooleanClause,
			values: []any{"true"},
			errStr: "invalid value for BooleanClause",
		},
		"NumberClause: int value rejected": {
			attr:   "age",
			op:     gtOp,
			kind:   NumberClause,
			values: []any{18},
			errStr: "invalid value for NumberClause",
		},
		"empty values rejected": {
			attr:   "name",
			op:     ilOp,
			kind:   StringClause,
			values: []any{},
			errStr: "can't create a clause with empty values",
		},
		"empty attribute rejected": {
			attr:   "",
			op:     eqOp,
			kind:   StringClause,
			values: []any{"a"},
			errStr: "can't create a clause with empty attribute",
		},
	}

	// Supported operators per kind are spelled out here rather than read from
	// clauseOperatorMap, so the test catches unintended changes to that map.
	kinds := []struct {
		kind    ClauseKind
		name    string
		attr    string
		value   any
		allowed []string
	}{
		{StringClause, "StringClause", "name", "Na", []string{eqOp, neqOp, stwOp, edwOp, cntOp, ilOp}},
		{BooleanClause, "BooleanClause", "isPro", true, []string{eqOp, neqOp}},
		{NumberClause, "NumberClause", "age", 18.0, []string{eqOp, neqOp, gtOp, ltOp, gteOp, lteOp, ilOp}},
	}

	for _, k := range kinds {
		for _, op := range operators {
			tc := newClauseTest{attr: k.attr, op: op, kind: k.kind, values: []any{k.value}}
			if slices.Contains(k.allowed, op) {
				tc.want = Clause{attribute: k.attr, operator: op, kind: k.kind, values: []any{k.value}}
				tests[fmt.Sprintf("%s: create with %s operator", k.name, op)] = tc
			} else {
				tc.errStr = fmt.Sprintf("unsupported operator for %s: %s", k.name, op)
				tests[fmt.Sprintf("%s: %s operator not supported", k.name, op)] = tc
			}
		}
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewClause(tc.attr, tc.op, tc.kind, tc.values)
			if tc.errStr != "" {
				if err == nil || err.Error() != tc.errStr {
					t.Fatalf("NewClause() error = %v, want %q", err, tc.errStr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClause() unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.want, got, cmp.AllowUnexported(Clause{}))
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
