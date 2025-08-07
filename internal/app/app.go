package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/config"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/kafka"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/logger"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/service"
)

type App struct {
	config     *config.Config
	logger     logger.Logger
	consumer   consumer.Consumer
	dlqService *service.DLQService
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log, err := logger.NewZapLogger(cfg.LogDevelopment)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	kafkaConsumer, err := kafka.NewKafkaConsumer(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	var dlqService *service.DLQService
	if cfg.KafkaTopicDlq != "" {
		dlqService, err = service.NewDLQService(
			[]string{cfg.KafkaBootstrapServers},
			cfg.KafkaTopicDlq,
			cfg.KafkaUsername,
			cfg.KafkaPassword,
		)
		if err != nil {
			log.Warnf("Failed to create DLQ service: %v", err)
		}
	}

	return &App{
		config:     cfg,
		logger:     log,
		consumer:   kafkaConsumer,
		dlqService: dlqService,
	}, nil
}

func (a *App) setupHandlers() error {
	orderProcessor := service.NewOrderProcessor(a.config.TargetServiceUrl, a.logger)

	err := a.consumer.Subscribe(a.config.KafkaTopic, orderProcessor.Process)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	orderValidator := &service.OrderValidator{}
	err = a.consumer.AddValidator(a.config.KafkaTopic, orderValidator)
	if err != nil {
		return fmt.Errorf("failed to add validator: %w", err)
	}

	return nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := a.setupHandlers()
	if err != nil {
		return fmt.Errorf("failed to setup handlers: %w", err)
	}

	err = a.consumer.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	a.logger.Infof("Application started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	a.logger.Infof("Shutting down application...")

	err = a.consumer.Stop(ctx)
	if err != nil {
		a.logger.Errorf("Error stopping consumer: %v", err)
	}

	if a.dlqService != nil {
		err = a.dlqService.Close()
		if err != nil {
			a.logger.Errorf("Error closing DLQ service: %v", err)
		}
	}

	a.logger.Infof("Application shutdown complete")
	return nil
}
