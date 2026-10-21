// Command server runs the calculator REST API.
//
// Run with -healthcheck to probe a running server and exit 0 (healthy) or 1.
// The Docker image uses this because the distroless base has no curl or shell.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/cmmmf/SezzleCalculator/backend/internal/config"
	"github.com/cmmmf/SezzleCalculator/backend/internal/controller"
	"github.com/cmmmf/SezzleCalculator/backend/internal/logger"
	"github.com/cmmmf/SezzleCalculator/backend/internal/router"
	"github.com/cmmmf/SezzleCalculator/backend/internal/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe GET /healthz on the local server and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}

	if *healthcheck {
		if err := probe(cfg.Port); err != nil {
			fmt.Fprintln(os.Stderr, "healthcheck:", err)
			os.Exit(1)
		}
		return
	}

	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cfg config.Config) error {
	log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	defer func() { _ = log.Sync() }() // Sync fails harmlessly on stdout/stderr.

	gin.SetMode(cfg.GinMode)
	engine := router.New(router.Options{
		Logger:      log,
		CORSOrigins: cfg.CORSOrigins,
		Calculator:  controller.NewCalculatorController(service.NewCalculatorService()),
	})

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr), zap.String("gin_mode", cfg.GinMode))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	log.Info("shutting down", zap.Duration("timeout", shutdownTimeout))
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Info("server stopped")
	return nil
}

func probe(port string) error {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
