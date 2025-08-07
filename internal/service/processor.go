package service

import (
	"context"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/logger"
)

type OrderProcessor struct {
}

func NewOrderProcessor(targetServiceUrl string, log logger.Logger) Processor {
	return &OrderProcessor{}
}

func (p *OrderProcessor) Process(ctx context.Context, msg models.Message) error {
	return nil
}
