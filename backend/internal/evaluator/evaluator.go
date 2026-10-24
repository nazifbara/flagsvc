package evaluator

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type ClauseKind int

const (
	StringClause ClauseKind = iota + 1
	BooleanClause
	NumberClause
)

const (
	eqOp  = "eq"
	neqOp = "neq"
	gtOp  = "gt"
	ltOp  = "lt"
	gteOp = "gte"
	lteOp = "lte"
	stwOp = "stw"
	edwOp = "edw"
	cntOp = "cnt"
	ilOp  = "il"
)

type Clause interface {
	Match(attributes map[string][]any) (bool, error)
}

type stringClause struct {
	attribute string
	operator  string
	values    []string
}

type numberClause struct {
	attribute string
	operator  string
	values    []float64
}

type booleanClause struct {
	attribute string
	operator  string
	values    []bool
}

func (c booleanClause) Match(attributes map[string][]any) (bool, error) {
	return false, nil
}

var operators = []string{eqOp, neqOp, gtOp, ltOp, gteOp, lteOp, stwOp, edwOp, cntOp, ilOp}

var clauseOperatorMap = map[ClauseKind][]string{
	StringClause:  {eqOp, neqOp, stwOp, edwOp, cntOp, ilOp},
	BooleanClause: {eqOp, neqOp},
	NumberClause:  {eqOp, neqOp, gtOp, ltOp, gteOp, lteOp, cntOp, ilOp},
}

func NewClause(attribute, operator string, kind ClauseKind, values []any) (Clause, error) {
	if attribute == "" {
		return nil, errors.New("can't create a clause with empty attribute")
	}
	if len(values) == 0 {
		return nil, errors.New("can't create a clause with empty values")
	}
	if !slices.Contains(operators, operator) {
		return nil, fmt.Errorf("unknown operator: %s", operator)
	}
	if slices.Contains([]string{eqOp, neqOp}, operator) && len(values) > 1 {
		return nil, fmt.Errorf("%v operator on multi-value clause", operator)
	}

	switch kind {
	case StringClause:
		strs, err := convert[string](values)
		if err != nil {
			return nil, fmt.Errorf("invalid value for StringClause: %w", err)
		}
		if !slices.Contains(clauseOperatorMap[StringClause], operator) {
			return nil, fmt.Errorf("unsupported operator for StringClause: %s", operator)
		}
		return stringClause{attribute: attribute, operator: operator, values: strs}, nil
	case BooleanClause:
		bools, err := convert[bool](values)
		if err != nil {
			return nil, fmt.Errorf("invalid value for BooleanClause: %w", err)
		}
		if !slices.Contains(clauseOperatorMap[BooleanClause], operator) {
			return nil, fmt.Errorf("unsupported operator for BooleanClause: %s", operator)
		}
		return booleanClause{attribute: attribute, operator: operator, values: bools}, nil
	case NumberClause:
		numbers, err := convert[float64](values)
		if err != nil {
			return nil, fmt.Errorf("invalid value for NumberClause: %w", err)
		}
		if !slices.Contains(clauseOperatorMap[NumberClause], operator) {
			return nil, fmt.Errorf("unsupported operator for NumberClause: %s", operator)
		}
		return numberClause{attribute: attribute, operator: operator, values: numbers}, nil
	default:
		return nil, errors.New("unknown clause kind")
	}
}

func convert[T any](values []any) ([]T, error) {
	out := make([]T, len(values))
	for i, v := range values {
		t, ok := v.(T)
		if !ok {
			return nil, fmt.Errorf("invalid value %v", v)
		}
		out[i] = t
	}
	return out, nil
}

func (c stringClause) Match(attributes map[string][]any) (bool, error) {
	values, ok := attributes[c.attribute]
	if !ok || len(values) == 0 {
		return false, nil
	}
	inputValues, err := convert[string](values)
	if err != nil {
		return false, fmt.Errorf("invalid value for attribute %q: %w", c.attribute, err)
	}
	switch c.operator {
	case ilOp:
		return valueContains(c.values, inputValues), nil
	case cntOp:
		return valueContains(inputValues, c.values), nil
	default:
		if len(values) > 1 {
			return false, fmt.Errorf("multi-value attribute not supported for %s operation in string clause", c.operator)
		}
		return matchString(c.operator, c.values[0], inputValues[0])
	}
}

func (c numberClause) Match(attributes map[string][]any) (bool, error) {
	values, ok := attributes[c.attribute]
	if !ok || len(values) == 0 {
		return false, nil
	}
	inputValues, err := convert[float64](values)
	if err != nil {
		return false, fmt.Errorf("invalid value for attribute %q: %w", c.attribute, err)
	}
	if !slices.Contains([]string{ilOp, cntOp}, c.operator) && len(inputValues) > 1 {
		return false, fmt.Errorf("multi-value attribute not supported for %s operation in number clause", c.operator)
	}
	switch c.operator {
	case ilOp:
		return valueContains(c.values, inputValues), nil
	case cntOp:
		return valueContains(inputValues, c.values), nil
	case eqOp:
		return inputValues[0] == c.values[0], nil
	case neqOp:
		return inputValues[0] != c.values[0], nil
	case gtOp:
		return inputValues[0] > c.values[0], nil
	case gteOp:
		return inputValues[0] >= c.values[0], nil
	case ltOp:
		return inputValues[0] < c.values[0], nil
	case lteOp:
		return inputValues[0] <= c.values[0], nil
	default:
		return false, fmt.Errorf("number clause with invalid operator: %q", c.operator)
	}
}

func valueContains[T comparable](a []T, b []T) bool {
	for _, v := range b {
		if !slices.Contains(a, v) {
			return false
		}
	}
	return true
}

func matchString(op, clauseValue, targetValue string) (bool, error) {
	switch op {
	case "eq":
		return clauseValue == targetValue, nil
	case "neq":
		return clauseValue != targetValue, nil
	case "stw":
		return strings.HasPrefix(targetValue, clauseValue), nil
	case "edw":
		return strings.HasSuffix(targetValue, clauseValue), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", op)
	}
}
