package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"time"

	store "github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal"
	"github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal/authservice"
	"github/nazifbara/sveltekit-go-connect-grpc-starter/gen/go/v1/sv1connect"

	connect "connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/validate"
)

type server struct {
	httpServer *http.Server
	cancel     context.CancelFunc
}

func (s *server) start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	log.Printf("server running http://localhost:%d", ln.Addr().(*net.TCPAddr).Port)
	if err := s.httpServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *server) shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func newServer(port int, store *store.Store, cancel context.CancelFunc) *server {
	connectServer := connect.NewServer(
		NewErrorHandlerInterceptor(),
		validate.NewServerInterceptor(),
	)
	connectServer.Register()
	sv1connect.RegisterAuthServiceHandler(
		connectServer,
		authservice.New(store),
	)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, connectServer)
	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
		Protocols:         p,
	}
	s := &server{
		httpServer: srv,
		cancel:     cancel,
	}
	return s
}

func NewErrorHandlerInterceptor() connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			sensitiveCodes := []connect.Code{connect.CodeInternal, connect.CodeAlreadyExists, connect.CodePermissionDenied}
			err := next(ctx, spec, stream)
			log.Println(err)
			if connectErr, ok := errors.AsType[*connect.Error](err); ok {
				if slices.Contains(sensitiveCodes, connectErr.Code()) {
					err = connect.NewError(connectErr.Code(), codeToText(connectErr.Code())).WithCause(err)
				}
			}
			return err
		}
	}
}

func codeToText(code connect.Code) string {
	var text string
	switch code {
	case connect.CodeAlreadyExists:
		text = "can't duplicate"
	case connect.CodeNotFound:
		text = "not found"
	default:
		text = "something went wrong"
	}
	return text
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
