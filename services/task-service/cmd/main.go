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

	authadapter "todoapp/services/task-service/internal/adapters/auth"
	dbadapter "todoapp/services/task-service/internal/adapters/database"
	"todoapp/services/task-service/internal/infrastructure/app"
	"todoapp/services/task-service/internal/infrastructure/config"
	"todoapp/services/task-service/internal/infrastructure/postgres"
	"todoapp/services/task-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repo := dbadapter.NewPostgresTaskRepository(pool)
	taskService := service.NewTaskService(repo)
	tokenManager := authadapter.NewJWTManager(cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	router, err := app.NewRouter(app.HTTPDeps{
		TaskService: taskService,
		TokenMgr:    tokenManager,
		ServiceName: cfg.ServiceName,
	})
	if err != nil {
		log.Fatalf("failed to initialize router: %v", err)
	}

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	<-done
}
