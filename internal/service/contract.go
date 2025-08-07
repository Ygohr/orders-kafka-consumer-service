package service

import (
	"context"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
)

type Processor interface {
	Process(ctx context.Context, msg models.Message) error
}
