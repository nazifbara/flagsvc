package authservice

import (
	"context"

	"connectrpc.com/connect/v2"
	"github.com/alexedwards/argon2id"

	"github/nazifbara/flagsvc/backend/internal/database"
	"github/nazifbara/flagsvc/backend/internal/omerrors"
	"github/nazifbara/flagsvc/backend/internal/store"
	auth "github/nazifbara/flagsvc/gen/go/flagsvc/auth/v1"
)

type Service struct {
	store *store.Store
}

func New(store *store.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) SignUp(
	ctx context.Context,
	req *auth.SignUpRequest,
) (*auth.SignUpResponse, error) {
	err := s.store.ExecTx(ctx, func(q *database.Queries) error {
		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
		}

		identity, err := q.CreateAuth(
			ctx,
			database.CreateAuthParams{
				Email:        req.Email,
				Passwordhash: passwordHash,
			},
		)
		if err != nil {
			return connect.NewError(connect.CodeAlreadyExists, err.Error()).WithCause(err)
		}

		_, err = q.CreateUser(
			ctx,
			database.CreateUserParams{
				ID:        identity.ID,
				Email:     req.Email,
				Firstname: req.FirstName,
				Lastname:  req.LastName,
			},
		)
		if err != nil {
			return omerrors.FromDBErr(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &auth.SignUpResponse{}, nil
}

func hashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
