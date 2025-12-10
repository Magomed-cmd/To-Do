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

	userv1 "todoapp/pkg/proto/user/v1"
	authadapter "todoapp/services/user-service/internal/adapters/auth"
	dbadapter "todoapp/services/user-service/internal/adapters/database"
	usergrpc "todoapp/services/user-service/internal/adapters/grpc"
	"todoapp/services/user-service/internal/infrastructure/app"
	"todoapp/services/user-service/internal/infrastructure/config"
	"todoapp/services/user-service/internal/infrastructure/posgres"
	"todoapp/services/user-service/internal/service"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := posgres.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	repo := dbadapter.NewPostgresUserRepository(pool)
	tokenManager := authadapter.NewJWTManager(cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	userService := service.NewUserService(repo, tokenManager)
	githubOAuth := authadapter.NewGitHubOAuth(cfg.GitHub.ClientID, cfg.GitHub.ClientSecret, cfg.GitHub.RedirectURL, cfg.GitHub.Scopes)

	router, err := app.NewRouter(app.HTTPDeps{
		UserService: userService,
		TokenMgr:    tokenManager,
		GitHubOAuth: githubOAuth,
	})
	if err != nil {
		log.Fatalf("failed to initialize router: %v", err)
	}

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", cfg.GRPCAddr, err)
	}
	defer grpcListener.Close()

	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, usergrpc.NewServer(userService, tokenManager))

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
		if err := grpcServer.Serve(grpcListener); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	grpcServer.GracefulStop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		log.Printf("server stopped with error: %v", err)
	case <-done:
	}
}
