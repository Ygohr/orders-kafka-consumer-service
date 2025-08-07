package consumer

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type MessageValidator interface {
	Validate(msg *sarama.ConsumerMessage) error
}

type DefaultValidator struct{}

func (v *DefaultValidator) Validate(msg *sarama.ConsumerMessage) error {
	if len(msg.Value) == 0 {
		return fmt.Errorf("message value is empty")
	}

	var rawPayload map[string]any
	if err := json.Unmarshal(msg.Value, &rawPayload); err != nil {
		return fmt.Errorf("message is not valid JSON: %w", err)
	}

	return nil
}
