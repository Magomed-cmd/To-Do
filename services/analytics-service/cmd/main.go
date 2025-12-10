package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	analyticsv1 "todoapp/pkg/proto/analytics/v1"
	dbadapter "todoapp/services/analytics-service/internal/adapters/database"
	grpcadapter "todoapp/services/analytics-service/internal/adapters/grpc"
	"todoapp/services/analytics-service/internal/infrastructure/app"
	"todoapp/services/analytics-service/internal/infrastructure/config"
	"todoapp/services/analytics-service/internal/infrastructure/postgres"
	"todoapp/services/analytics-service/internal/service"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := log.New(os.Stdout, "analytics-service ", log.LstdFlags|log.Lshortfile)

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		logger.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repo := dbadapter.NewPostgresRepository(pool)
	analyticsService := service.New(repo)

	router, err := app.NewRouter(app.HTTPDeps{ServiceName: cfg.ServiceName, Analytics: analyticsService})
	if err != nil {
		logger.Fatalf("failed to initialize http router: %v", err)
	}

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: router}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		logger.Fatalf("failed to listen on %s: %v", cfg.GRPCAddr, err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	analyticsv1.RegisterAnalyticsServiceServer(grpcServer, grpcadapter.NewServer(analyticsService))

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Printf("http shutdown error: %v", err)
	}

	grpcServer.GracefulStop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		logger.Printf("server stopped with error: %v", err)
	case <-done:
	}
}
