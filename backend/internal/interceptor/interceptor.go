package interceptor

import (
	"context"
	"errors"
	"log"
	"slices"

	"github/nazifbara/flagsvc/backend/internal/utils"

	connect "connectrpc.com/connect/v2"
)

func NewLoggingInterceptor() connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			sensitiveCodes := []connect.Code{connect.CodeInternal, connect.CodeAlreadyExists, connect.CodePermissionDenied}
			err := next(ctx, spec, stream)
			if err != nil {
				log.Println(err)
			}
			if connectErr, ok := errors.AsType[*connect.Error](err); ok {
				if slices.Contains(sensitiveCodes, connectErr.Code()) {
					err = connect.NewError(connectErr.Code(), utils.CodeToText(connectErr.Code())).WithCause(err)
				}
			}
			return err
		}
	}
}
