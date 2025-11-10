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

	authadapter "todoapp/services/user-service/internal/adapters/auth"
	dbadapter "todoapp/services/user-service/internal/adapters/database"
	"todoapp/services/user-service/internal/infrastructure/app"
	"todoapp/services/user-service/internal/infrastructure/config"
	"todoapp/services/user-service/internal/infrastructure/posgres"
	"todoapp/services/user-service/internal/service"
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

