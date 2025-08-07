package service

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestOrderValidator_Validate(t *testing.T) {
	validator := &OrderValidator{}

	tests := []struct {
		name    string
		message *sarama.ConsumerMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid order payload",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"ordemDeVenda": "order123", "etapaAtual": "FATURADO"}`),
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
			name: "missing ordemDeVenda field",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"etapaAtual": "FATURADO"}`),
			},
			wantErr: true,
			errMsg:  "required field 'ordemDeVenda' is missing",
		},
		{
			name: "missing etapaAtual field",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"ordemDeVenda": "order123"}`),
			},
			wantErr: true,
			errMsg:  "required field 'etapaAtual' is missing",
		},
		{
			name: "empty ordemDeVenda value",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"ordemDeVenda": "", "etapaAtual": "FATURADO"}`),
			},
			wantErr: false,
		},
		{
			name: "empty etapaAtual value",
			message: &sarama.ConsumerMessage{
				Value: []byte(`{"ordemDeVenda": "order123", "etapaAtual": ""}`),
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
