package consumer

import (
	"context"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
)

type MessageHandler func(ctx context.Context, msg models.Message) error

type Consumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Subscribe(topic string, handler MessageHandler) error
	AddValidator(topic string, validator MessageValidator) error
}
