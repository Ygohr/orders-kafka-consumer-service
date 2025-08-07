package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/logger"
)

type OrderProcessor struct {
	targetServiceUrl string
	log              logger.Logger
	httpClient       *http.Client
}

func NewOrderProcessor(targetServiceUrl string, log logger.Logger) Processor {
	return &OrderProcessor{
		targetServiceUrl: targetServiceUrl,
		log:              log,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *OrderProcessor) Process(ctx context.Context, msg models.Message) error {
	p.log.Infof("Processing message - topic: %s, partition: %d, offset: %d", msg.Topic, msg.Partition, msg.Offset)

	var payload models.OrderPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		p.log.Errorf("Failed to unmarshal payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.log.Infof("Payload parsed successfully - ordemDeVenda: %s, etapaAtual: %s", payload.OrdemDeVenda, payload.EtapaAtual)

	if err := p.callTargetService(ctx, payload); err != nil {
		p.log.Errorf("Failed to forward payload to target service: %v", err)
		return fmt.Errorf("failed to forward payload: %w", err)
	}

	p.log.Infof("Message processed successfully - ordemDeVenda: %s", payload.OrdemDeVenda)

	return nil
}

func (p *OrderProcessor) callTargetService(ctx context.Context, payload models.OrderPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", p.targetServiceUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	p.log.Infof("Sending request to target service - url: %s, payload: %s", p.targetServiceUrl, string(payloadBytes))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("target service returned error status: %d", resp.StatusCode)
	}

	p.log.Infof("Successfully forwarded payload to target service - status: %d", resp.StatusCode)

	return nil
}
