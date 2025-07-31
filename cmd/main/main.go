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
	"github.com/supchaser/test_task/internal/app/delivery"
	"github.com/supchaser/test_task/internal/app/repository"
	"github.com/supchaser/test_task/internal/app/usecase"
	"github.com/supchaser/test_task/internal/config"
	"github.com/supchaser/test_task/internal/middleware"
	"github.com/supchaser/test_task/internal/utils/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig(".env")
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
	logger.Debug("debug mode enabled",
		zap.String("log_mode", cfg.LogMode),
		zap.Int("max_tasks", cfg.MaxActiveTasks),
	)

	if err := os.MkdirAll("./storage", 0755); err != nil {
		logger.Error("failed to create storage directory", zap.Error(err))
		os.Exit(1)
	}

	taskRepo := repository.CreateTaskRepository(cfg.MaxActiveTasks)
	taskUsecase := usecase.CreateTaskUsecase(taskRepo, "")
	taskDelivery := delivery.CreateTaskDelivery(taskUsecase)

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	taskRouter := apiRouter.PathPrefix("/tasks").Subrouter()
	taskRouter.HandleFunc("", taskDelivery.CreateTask).Methods("POST")
	taskRouter.HandleFunc("", taskDelivery.GetAllTasks).Methods("GET")
	taskRouter.HandleFunc("/{id:[0-9]+}", taskDelivery.GetTask).Methods("GET")
	taskRouter.HandleFunc("/{id:[0-9]+}/objects", taskDelivery.AddObjects).Methods("POST")
	taskRouter.HandleFunc("/{id:[0-9]+}/archive", taskDelivery.DownloadArchive).Methods("GET")
	taskRouter.HandleFunc("/{id:[0-9]+}/status", taskDelivery.GetTaskStatus).Methods("GET")

	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.PanicMiddleware)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting HTTP server",
			zap.String("address", server.Addr),
			zap.Any("config", cfg),
		)
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
		logger.Info("server is shutting down",
			zap.String("signal", sig.String()),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
			os.Exit(1)
		}

		logger.Info("server stopped")
	}
}
