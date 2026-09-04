package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VaPsDani/sezzle-calculator/backend/internal/api"
)

const (
	defaultPort       = "8080"
	readTimeout       = 5 * time.Second
	readHeaderTimeout = 2 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           api.NewRouter(),
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverFailed := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverFailed <- err
		}
	}()

	select {
	case err := <-serverFailed:
		log.Fatalf("server error: %v", err)

	case <-ctx.Done():
		// Restores the default behaviour so a second Ctrl+C during the drain
		// kills the process instead of being swallowed.
		stop()
		log.Println("shutdown signal received, draining in flight requests")
	}

	// ctx is already cancelled, so the drain needs a deadline of its own.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown timed out: %v", err)
		if closeErr := srv.Close(); closeErr != nil {
			log.Printf("forced close failed: %v", closeErr)
		}
		os.Exit(1)
	}

	log.Println("server stopped cleanly")
}
