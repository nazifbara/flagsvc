package evaluator

import (
	"errors"
	"fmt"
	"slices"
)

type ClauseKind int

const (
	StringClause ClauseKind = iota
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

type Clause struct {
	kind      ClauseKind
	attribute string
	operator  string
	values    []any
}

var operators = []string{eqOp, neqOp, gtOp, ltOp, gteOp, lteOp, stwOp, edwOp, cntOp, ilOp}

var clauseOperatorMap = map[ClauseKind][]string{
	StringClause:  {eqOp, neqOp, stwOp, edwOp, cntOp, ilOp},
	BooleanClause: {eqOp, neqOp},
	NumberClause:  {eqOp, neqOp, gtOp, ltOp, gteOp, lteOp, cntOp, ilOp},
}

func NewClause(attr, operator string, kind ClauseKind, values []any) (Clause, error) {
	if attr == "" {
		return Clause{}, errors.New("can't create a clause with empty attribute")
	}
	if len(values) == 0 {
		return Clause{}, errors.New("can't create a clause with empty values")
	}
	if !slices.Contains(operators, operator) {
		return Clause{}, fmt.Errorf("unknown operator: %s", operator)
	}
	if slices.Contains([]string{eqOp, neqOp}, operator) && len(values) > 1 {
		return Clause{}, fmt.Errorf("%v operator on multi-value clause", operator)
	}

	switch kind {
	case StringClause:
		for _, value := range values {
			if _, ok := value.(string); !ok {
				return Clause{}, errors.New("invalid value for StringClause")
			}
		}
		if !slices.Contains(clauseOperatorMap[StringClause], operator) {
			return Clause{}, fmt.Errorf("unsupported operator for StringClause: %s", operator)
		}
	case BooleanClause:
		for _, value := range values {
			if _, ok := value.(bool); !ok {
				return Clause{}, errors.New("invalid value for BooleanClause")
			}
		}
		if !slices.Contains(clauseOperatorMap[BooleanClause], operator) {
			return Clause{}, fmt.Errorf("unsupported operator for BooleanClause: %s", operator)
		}
	case NumberClause:
		for _, value := range values {
			if _, ok := value.(float64); !ok {
				return Clause{}, errors.New("invalid value for NumberClause")
			}
		}
		if !slices.Contains(clauseOperatorMap[NumberClause], operator) {
			return Clause{}, fmt.Errorf("unsupported operator for NumberClause: %s", operator)
		}
	}

	return Clause{
		kind:      kind,
		attribute: attr,
		operator:  operator,
		values:    values,
	}, nil
}
