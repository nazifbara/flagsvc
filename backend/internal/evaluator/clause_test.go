package evaluator

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewClause(t *testing.T) {
	type newClauseTest struct {
		attr   string
		op     string
		kind   ClauseKind
		values []any
		want   Clause
		errStr string
	}
	tests := map[string]newClauseTest{
		"unknown operator": {
			attr:   "abc",
			op:     "equal",
			kind:   BooleanClause,
			values: []any{true},
			errStr: "unknown operator: equal",
		},
		"unknown kind rejected": {
			attr:   "abc",
			op:     eqOp,
			kind:   10000,
			values: []any{"a"},
			errStr: "unknown clause kind",
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
			want:   stringClause{attribute: "country", operator: ilOp, values: []string{"BJ", "FR"}},
		},
		"cnt operator accepts multiple values": {
			attr:   "country",
			op:     cntOp,
			kind:   StringClause,
			values: []any{"BJ", "FR"},
			want:   stringClause{attribute: "country", operator: cntOp, values: []string{"BJ", "FR"}},
		},
		"StringClause: invalid value type": {
			attr:   "abc",
			op:     ilOp,
			kind:   StringClause,
			values: []any{"1", 2, 3},
			errStr: "invalid value for StringClause: invalid value 2",
		},
		"StringClause: invalid value checked before unsupported operator": {
			attr:   "role",
			op:     gtOp,
			kind:   StringClause,
			values: []any{1.5},
			errStr: "invalid value for StringClause: invalid value 1.5",
		},
		"BooleanClause: invalid value type": {
			attr:   "isPro",
			op:     eqOp,
			kind:   BooleanClause,
			values: []any{"true"},
			errStr: "invalid value for BooleanClause: invalid value true",
		},
		"NumberClause: int value rejected": {
			attr:   "age",
			op:     gtOp,
			kind:   NumberClause,
			values: []any{18},
			errStr: "invalid value for NumberClause: invalid value 18",
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
		value   any
		allowed []string
		want    func(op string) Clause
	}{
		{
			StringClause, "StringClause", "Na",
			[]string{eqOp, neqOp, stwOp, edwOp, cntOp, ilOp},
			func(op string) Clause { return stringClause{attribute: "attr", operator: op, values: []string{"Na"}} },
		},
		{
			BooleanClause, "BooleanClause", true,
			[]string{eqOp, neqOp},
			func(op string) Clause { return booleanClause{attribute: "attr", operator: op, values: []bool{true}} },
		},
		{
			NumberClause, "NumberClause", 18.0,
			[]string{eqOp, neqOp, gtOp, ltOp, gteOp, lteOp, ilOp, cntOp},
			func(op string) Clause { return numberClause{attribute: "attr", operator: op, values: []float64{18.0}} },
		},
	}

	for _, k := range kinds {
		for _, op := range operators {
			tc := newClauseTest{attr: "attr", op: op, kind: k.kind, values: []any{k.value}}
			if slices.Contains(k.allowed, op) {
				tc.want = k.want(op)
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
				if got != nil {
					t.Fatalf("NewClause() = %#v on error, want nil", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClause() unexpected error: %v", err)
			}

			opt := cmp.AllowUnexported(stringClause{}, booleanClause{}, numberClause{})
			if diff := cmp.Diff(tc.want, got, opt); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func mustNewClause(t *testing.T, attr, op string, kind ClauseKind, values ...any) Clause {
	t.Helper()
	c, err := NewClause(attr, op, kind, values)
	if err != nil {
		t.Fatalf("NewClause(%q, %q, %v, %v): %v", attr, op, kind, values, err)
	}
	return c
}

func TestStringClauseMatch(t *testing.T) {
	str := func(op string, values ...any) Clause {
		return mustNewClause(t, "country", op, StringClause, values...)
	}

	tests := map[string]struct {
		clause Clause
		input  map[string][]any
		want   bool
		errStr string
	}{
		"false when attribute is not found": {
			clause: str(eqOp, "BJ"),
			input:  map[string][]any{"language": {"FR"}},
			want:   false,
		},
		"neq is false when attribute is not found": {
			clause: str(neqOp, "US"),
			input:  map[string][]any{"language": {"FR"}},
			want:   false,
		},
		"false on empty attribute values": {
			clause: str(eqOp, "BJ"),
			input:  map[string][]any{"country": {}},
			want:   false,
		},
		"non-string attribute value rejected": {
			clause: str(eqOp, "BJ"),
			input:  map[string][]any{"country": {1.0}},
			errStr: `invalid value for attribute "country": invalid value 1`,
		},
		"eq rejects attribute with multiple values": {
			clause: str(eqOp, "US"),
			input:  map[string][]any{"country": {"BJ", "US"}},
			errStr: "multi-value attribute not supported for eq operation in string clause",
		},
		"neq rejects attribute with multiple values": {
			clause: str(neqOp, "US"),
			input:  map[string][]any{"country": {"BJ", "US"}},
			errStr: "multi-value attribute not supported for neq operation in string clause",
		},
		"stw rejects attribute with multiple values": {
			clause: str(stwOp, "B"),
			input:  map[string][]any{"country": {"BJ", "BE"}},
			errStr: "multi-value attribute not supported for stw operation in string clause",
		},
		"edw rejects attribute with multiple values": {
			clause: str(edwOp, "J"),
			input:  map[string][]any{"country": {"BJ", "NJ"}},
			errStr: "multi-value attribute not supported for edw operation in string clause",
		},
		"eq true when equal": {
			clause: str(eqOp, "BJ"),
			input:  map[string][]any{"country": {"BJ"}, "language": {"FR"}},
			want:   true,
		},
		"eq false when different": {
			clause: str(eqOp, "US"),
			input:  map[string][]any{"country": {"BJ"}, "language": {"FR"}},
			want:   false,
		},
		"eq is case-sensitive": {
			clause: str(eqOp, "bj"),
			input:  map[string][]any{"country": {"BJ"}},
			want:   false,
		},
		"neq true when different": {
			clause: str(neqOp, "US"),
			input:  map[string][]any{"country": {"BJ"}, "language": {"FR"}},
			want:   true,
		},
		"neq false when equal": {
			clause: str(neqOp, "BJ"),
			input:  map[string][]any{"country": {"BJ"}, "language": {"FR"}},
			want:   false,
		},
		"il true when every attribute value is in the list": {
			clause: str(ilOp, "BJ", "NG", "US"),
			input:  map[string][]any{"country": {"US", "BJ"}},
			want:   true,
		},
		"il false when no attribute value is in the list": {
			clause: str(ilOp, "BJ", "NG", "US"),
			input:  map[string][]any{"country": {"NZ", "FR"}},
			want:   false,
		},
		"il false when only some attribute values are in the list": {
			clause: str(ilOp, "BJ", "NG", "US"),
			input:  map[string][]any{"country": {"US", "NZ"}},
			want:   false,
		},
		"cnt true when attribute has every clause value": {
			clause: str(cntOp, "BJ", "NG"),
			input:  map[string][]any{"country": {"NZ", "BJ", "FR", "NG"}},
			want:   true,
		},
		"cnt false when attribute misses a clause value": {
			clause: str(cntOp, "BJ", "NG"),
			input:  map[string][]any{"country": {"NZ", "BJ", "FR"}},
			want:   false,
		},
		// Same clause values and input: il and cnt must disagree.
		"il true where cnt is false": {
			clause: str(ilOp, "BJ", "FR", "US"),
			input:  map[string][]any{"country": {"BJ", "FR"}},
			want:   true,
		},
		"cnt false where il is true": {
			clause: str(cntOp, "BJ", "FR", "US"),
			input:  map[string][]any{"country": {"BJ", "FR"}},
			want:   false,
		},
		"stw true on prefix": {
			clause: str(stwOp, "he"),
			input:  map[string][]any{"country": {"hello world"}},
			want:   true,
		},
		"stw true when prefix is the whole value": {
			clause: str(stwOp, "hello"),
			input:  map[string][]any{"country": {"hello"}},
			want:   true,
		},
		"stw false when not a prefix": {
			clause: str(stwOp, "world"),
			input:  map[string][]any{"country": {"hello world"}},
			want:   false,
		},
		"stw is case-sensitive": {
			clause: str(stwOp, "He"),
			input:  map[string][]any{"country": {"hello"}},
			want:   false,
		},
		"edw true on suffix": {
			clause: str(edwOp, "rld"),
			input:  map[string][]any{"country": {"hello world"}},
			want:   true,
		},
		"edw false when not a suffix": {
			clause: str(edwOp, "rld"),
			input:  map[string][]any{"country": {"let's go touch some grass ld"}},
			want:   false,
		},
		// Not reachable through NewClause; guards checkString's default branch.
		"unsupported operator on hand-built clause": {
			clause: stringClause{attribute: "country", operator: gtOp, values: []string{"BJ"}},
			input:  map[string][]any{"country": {"BJ"}},
			errStr: "unknown operator: gt",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tc.clause.Match(tc.input)
			if tc.errStr != "" {
				if err == nil || err.Error() != tc.errStr {
					t.Fatalf("Match() error = %v, want %q", err, tc.errStr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Match() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Match() = %v, want %v", got, tc.want)
			}
		})
	}
}
