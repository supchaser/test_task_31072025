package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/supchaser/test_task/internal/config"
	"github.com/supchaser/test_task/internal/utils/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("error initializing config: %v\n", err)
		os.Exit(1)
	}

	err = logger.Init(cfg.LogMode)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("configuration loaded successfully")
	logger.Debug("debug mode enabled", zap.String("log_mode", cfg.LogMode))

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("failed to start server", zap.Error(err))
		os.Exit(1)
	case sig := <-quit:
		logger.Info("server is shutting down", zap.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
			os.Exit(1)
		}

		logger.Info("server stopped")
	}
}
