package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	rabbitmq "todoapp/services/notification-service/internal/adapters/rabbitmq"
	smtpadapter "todoapp/services/notification-service/internal/adapters/smtp"
	"todoapp/services/notification-service/internal/adapters/templates"
	"todoapp/services/notification-service/internal/infrastructure/config"
	"todoapp/services/notification-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := log.New(os.Stdout, "notification-service ", log.LstdFlags|log.Lshortfile)

	templateEngine, err := templates.NewEngine()
	if err != nil {
		log.Fatalf("template error: %v", err)
	}

	mailer := smtpadapter.NewMailer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From)
	notificationService := service.NewNotificationService(mailer, templateEngine)
	consumer := rabbitmq.NewConsumer(cfg.Rabbit.URL, cfg.Rabbit.Queue, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Println("shutdown signal received")
		cancel()
	}()

	if err := consumer.Consume(ctx, notificationService); err != nil {
		logger.Fatalf("consumer stopped: %v", err)
	}
}
