package authservice

import (
	"context"

	store "github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal"
	"github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal/database"
	"github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal/omerrors"
	sv1 "github/nazifbara/sveltekit-go-connect-grpc-starter/gen/go/v1"

	"connectrpc.com/connect"
	"github.com/alexedwards/argon2id"
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
	req *connect.Request[sv1.SignUpRequest],
) (*connect.Response[sv1.SignUpResponse], error) {
	err := s.store.ExecTx(ctx, func(q *database.Queries) error {
		passwordHash, err := hashPassword(req.Msg.Password)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		identity, err := q.CreateAuth(
			ctx,
			database.CreateAuthParams{
				Email:        req.Msg.Email,
				Passwordhash: passwordHash,
			},
		)
		if err != nil {
			return connect.NewError(connect.CodeAlreadyExists, err)
		}

		_, err = q.CreateUser(
			ctx,
			database.CreateUserParams{
				UserID:    identity.ID,
				Email:     req.Msg.Email,
				Firstname: req.Msg.FirstName,
				Lastname:  req.Msg.LastName,
				Timezone:  req.Msg.Timezone,
			},
		)
		if err != nil {
			return omerrors.FromDBErr(err)
		}
		return nil
	})
	if err != nil {
		return &connect.Response[sv1.SignUpResponse]{}, err
	}
	return &connect.Response[sv1.SignUpResponse]{}, nil
}

func hashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
