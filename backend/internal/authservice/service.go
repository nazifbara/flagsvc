package authservice

import (
	"context"

	"connectrpc.com/connect/v2"
	"github.com/alexedwards/argon2id"

	store "github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal"
	"github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal/database"
	"github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal/omerrors"
	sv1 "github/nazifbara/sveltekit-go-connect-grpc-starter/gen/go/v1"
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
	req *sv1.SignUpRequest,
) (*sv1.SignUpResponse, error) {
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
				UserID:    identity.ID,
				Email:     req.Email,
				Firstname: req.FirstName,
				Lastname:  req.LastName,
				Timezone:  req.Timezone,
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

	return &sv1.SignUpResponse{}, nil
}

func hashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
