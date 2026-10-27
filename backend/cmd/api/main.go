// package main is the program entry point. `go run ./cmd/api` builds this file
// and calls main. Other folders under internal/ are libraries, not programs.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myseat/internal/platform/config"
	"myseat/internal/platform/httpapi"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// := declares a variable and assigns it in one step. The type is inferred.
	// slog writes structured logs. nil means "use the default handler options".
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	// & allocates a struct and gives us a pointer to it.
	// Named fields (Addr:, Handler:) set only those fields; the rest stay at zero.
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httpapi.NewRouter(log),
		// Close connections that send headers too slowly, so one client cannot hold a slot open.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// A function can return more than one value. ctx becomes done when the
	// process receives Ctrl+C (SIGINT) or a termination signal (SIGTERM).
	// defer runs stop() when main returns, which releases the signal listener.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// A channel passes a value from one goroutine to another.
	// make(chan error, 1) is a mailbox with room for one error.
	errCh := make(chan error, 1)

	// go starts the function on a new goroutine, then main keeps going.
	// ListenAndServe blocks, so it must not run on the main goroutine or we
	// could never watch for Ctrl+C at the same time.
	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr)
		// <- sends the return value into errCh. ListenAndServe returns only when it stops.
		errCh <- srv.ListenAndServe()
	}()

	// select waits until one case can proceed, then runs that case.
	// It is a wait on several events, not a switch on a value we already have.
	select {
	// <-errCh receives from the channel. This case runs if the server stops on its own.
	case err := <-errCh:
		// errors.Is checks the error chain. ErrServerClosed is the normal result
		// after Shutdown, so only any other error is a failure.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped", "err", err)
			os.Exit(1)
		}
	// ctx.Done() is a channel that closes when the signal arrives.
	case <-ctx.Done():
		log.Info("shutting down")
		// WithTimeout gives Shutdown a deadline. cancel must run to free the timer.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		// Shutdown stops new requests and waits for the ones already running.
		// `if err := ...; err != nil` declares err only for this if statement.
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown", "err", err)
			os.Exit(1)
		}
	}
}
