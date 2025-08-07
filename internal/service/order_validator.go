package service

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type OrderValidator struct{}

func (v *OrderValidator) Validate(msg *sarama.ConsumerMessage) error {
	if len(msg.Value) == 0 {
		return fmt.Errorf("message value is empty")
	}

	var rawPayload map[string]any
	if err := json.Unmarshal(msg.Value, &rawPayload); err != nil {
		return fmt.Errorf("message is not valid JSON: %w", err)
	}

	if _, exists := rawPayload["ordemDeVenda"]; !exists {
		return fmt.Errorf("required field 'ordemDeVenda' is missing")
	}

	if _, exists := rawPayload["etapaAtual"]; !exists {
		return fmt.Errorf("required field 'etapaAtual' is missing")
	}

	return nil
}
