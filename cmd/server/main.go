// Command server is the entrypoint for go-farm-production. It bootstraps the
// application, starts the HTTP server, and performs graceful shutdown on
// SIGINT/SIGTERM.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"go-farm-production/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := app.Bootstrap(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	// Listen for termination signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in the background; exit on signal or server error.
	errCh := make(chan error, 1)
	go func() {
		if err := a.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			a.Logger().Error("server error", zap.Error(err))
			os.Exit(1)
		}
	case sig := <-sigCh:
		a.Logger().Info("signal received, shutting down", zap.String("signal", sig.String()))
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelShutdown()
		if err := a.Shutdown(shutdownCtx); err != nil {
			a.Logger().Error("graceful shutdown failed", zap.Error(err))
			os.Exit(1)
		}
	}
}
