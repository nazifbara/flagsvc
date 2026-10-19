package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/nazifbara/flagsvc/backend/internal/authservice"
	"github.com/nazifbara/flagsvc/backend/internal/interceptor"
	"github.com/nazifbara/flagsvc/backend/internal/pingservice"
	"github.com/nazifbara/flagsvc/backend/internal/store"
	"github.com/nazifbara/flagsvc/gen/go/flagsvc/auth/v1/authconnect"
	"github.com/nazifbara/flagsvc/gen/go/flagsvc/ping/v1/pingconnect"

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
		interceptor.NewLoggingInterceptor(),
		validate.NewServerInterceptor(),
	)
	authconnect.RegisterAuthServiceHandler(
		connectServer,
		authservice.New(store),
	)
	pingconnect.RegisterPingServiceHandler(
		connectServer,
		pingservice.PingService{},
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
