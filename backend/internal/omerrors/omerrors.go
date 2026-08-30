// Package omerrors maps errors to connect.Error
package omerrors

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/lib/pq"
)

func FromDBErr(err error) error {
	var newErr error
	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		switch pqErr.Code {
		case "23505":
			newErr = connect.NewError(connect.CodeAlreadyExists, err)
		default:
			newErr = connect.NewError(connect.CodeInternal, err)
		}
	}
	return newErr
}
