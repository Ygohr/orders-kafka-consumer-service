package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/config"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/service"
)

type KafkaConsumer struct {
	consumer     sarama.ConsumerGroup
	config       *config.Config
	handlers     map[string]consumer.MessageHandler
	running      bool
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	groupHandler *consumerGroupHandler
	dlqService   *service.DLQService
}

type consumerGroupHandler struct {
	handlers   map[string]consumer.MessageHandler
	mu         sync.RWMutex
	dlqService *service.DLQService
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session setup")
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session cleanup")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume claims for topic: %s, partition: %d", claim.Topic(), claim.InitialOffset())

	for {
		select {
		case <-session.Context().Done():
			log.Printf("Session context done for topic: %s", claim.Topic())
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Printf("Message channel closed for topic: %s", claim.Topic())
				return nil
			}

			log.Printf("Received message from topic: %s, partition: %d, offset: %d",
				msg.Topic, msg.Partition, msg.Offset)

			message := models.Message{
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       msg.Key,
				Value:     msg.Value,
				Headers:   make(map[string][]byte),
			}

			for _, header := range msg.Headers {
				message.Headers[string(header.Key)] = header.Value
			}

			h.mu.RLock()
			handler, exists := h.handlers[message.Topic]
			h.mu.RUnlock()

			if exists {
				if err := h.processMessageWithRetry(session.Context(), message, handler); err != nil {
					log.Printf("Failed to process message after retries from topic %s: %v", message.Topic, err)
					if h.dlqService != nil {
						if dlqErr := h.dlqService.SendToDLQ(session.Context(), message, err, 3); dlqErr != nil {
							log.Printf("Failed to send message to DLQ: %v", dlqErr)
						}
					}
				} else {
					log.Printf("Successfully processed message from topic: %s", message.Topic)
				}
			} else {
				log.Printf("No handler found for topic: %s", message.Topic)
			}

			session.MarkMessage(msg, "")
		}
	}
}

func (h *consumerGroupHandler) processMessageWithRetry(ctx context.Context, message models.Message, handler consumer.MessageHandler) error {
	maxRetries := 3
	backoffDelay := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := handler(ctx, message)
		if err == nil {
			return nil
		}

		log.Printf("Attempt %d/%d failed for message from topic %s: %v", attempt, maxRetries, message.Topic, err)

		if attempt == maxRetries {
			return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffDelay):
			backoffDelay *= 2
		}
	}

	return fmt.Errorf("unexpected error in retry logic")
}

func NewKafkaConsumer(cfg *config.Config) (*KafkaConsumer, error) {
	saramaConfig := sarama.NewConfig()

	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	saramaConfig.Net.DialTimeout = 10 * time.Second
	saramaConfig.Net.ReadTimeout = 10 * time.Second
	saramaConfig.Net.WriteTimeout = 10 * time.Second

	if cfg.KafkaUsername != "" && cfg.KafkaPassword != "" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		saramaConfig.Net.SASL.User = cfg.KafkaUsername
		saramaConfig.Net.SASL.Password = cfg.KafkaPassword
		saramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
		}
	}

	saramaConfig.Version = sarama.V2_8_0_0

	consumerGroup, err := sarama.NewConsumerGroup([]string{cfg.KafkaBootstrapServers}, cfg.KafkaGroupId, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer group: %w", err)
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
			log.Printf("Warning: Failed to create DLQ service: %v", err)
		} else {
			log.Printf("DLQ service initialized for topic: %s", cfg.KafkaTopicDlq)
		}
	}

	groupHandler := &consumerGroupHandler{
		handlers:   make(map[string]consumer.MessageHandler),
		dlqService: dlqService,
	}

	return &KafkaConsumer{
		consumer:     consumerGroup,
		config:       cfg,
		handlers:     make(map[string]consumer.MessageHandler),
		groupHandler: groupHandler,
		dlqService:   dlqService,
	}, nil
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if kc.running {
		return fmt.Errorf("consumer already running")
	}

	kc.ctx, kc.cancel = context.WithCancel(ctx)
	kc.running = true

	kc.groupHandler.mu.Lock()
	for topic, handler := range kc.handlers {
		kc.groupHandler.handlers[topic] = handler
	}
	kc.groupHandler.mu.Unlock()

	topics := make([]string, 0, len(kc.handlers))
	for topic := range kc.handlers {
		topics = append(topics, topic)
	}

	log.Printf("Starting consumer for topics: %v", topics)

	go func() {
		for {
			select {
			case <-kc.ctx.Done():
				log.Printf("Consumer context done, stopping consumer group")
				return
			default:
				if err := kc.consumer.Consume(kc.ctx, topics, kc.groupHandler); err != nil {
					log.Printf("Error from consumer: %v", err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	log.Printf("Kafka consumer started successfully")
	return nil
}

func (kc *KafkaConsumer) Stop(ctx context.Context) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if !kc.running {
		return nil
	}

	kc.running = false
	if kc.cancel != nil {
		kc.cancel()
	}

	if err := kc.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close kafka consumer: %w", err)
	}

	if kc.dlqService != nil {
		if err := kc.dlqService.Close(); err != nil {
			log.Printf("Warning: Failed to close DLQ service: %v", err)
		}
	}

	log.Printf("Kafka consumer stopped successfully")
	return nil
}

func (kc *KafkaConsumer) Subscribe(topic string, handler consumer.MessageHandler) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	kc.handlers[topic] = handler
	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

func (kc *KafkaConsumer) IsRunning() bool {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	return kc.running
}
