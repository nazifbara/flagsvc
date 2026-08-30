package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	store "github/nazifbara/sveltekit-go-connect-grpc-starter/backend/internal"

	_ "github.com/lib/pq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	httpPort := flag.Int("port", 8080, "port to listen on")
	flag.Parse()
	status := run(ctx, cancel, *httpPort)
	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc, httpPort int) int {
	store, err := store.New(os.Getenv("DB_URL"))
	if err != nil {
		log.Printf("failed to connect to db: %v", err)
		return 1
	}
	s := newServer(httpPort, store, cancel)
	var serverErr error
	go func() {
		serverErr = s.start()
		cancel()
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("the server is shutting down")
	if err := s.shutdown(shutdownCtx); err != nil {
		log.Printf("failed to shudown the server %v", err)
		return 1
	}
	if serverErr != nil {
		log.Printf("server error: %v", serverErr)
	}
	return 0
}
