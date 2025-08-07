package consumer

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestDefaultValidator_Validate(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name    string
		message *sarama.ConsumerMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid JSON message",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"key": "value"}`),
			},
			wantErr: false,
		},
		{
			name: "empty message value",
			message: &sarama.ConsumerMessage{
				Value: []byte{},
			},
			wantErr: true,
			errMsg:  "message value is empty",
		},
		{
			name: "invalid JSON",
			message: &sarama.ConsumerMessage{
				Value: []byte(`invalid json`),
			},
			wantErr: true,
			errMsg:  "message is not valid JSON",
		},
		{
			name: "null message value",
			message: &sarama.ConsumerMessage{
				Value: nil,
			},
			wantErr: true,
			errMsg:  "message value is empty",
		},
		{
			name: "complex JSON object",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"nested": {"key": "value"}, "array": [1, 2, 3]}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.message)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
