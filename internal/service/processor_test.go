package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
	"github.com/Ygohr/orders-kafka-consumer-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOrderProcessor(t *testing.T) {
	mockLogger := &mocks.MockLogger{}
	targetURL := "http://localhost:8081/api/v1/orders"

	processor := NewOrderProcessor(targetURL, mockLogger)

	assert.NotNil(t, processor)
}

func TestOrderProcessor_Process_ValidMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload models.OrderPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "order123", payload.OrdemDeVenda)
		assert.Equal(t, "FATURADO", payload.EtapaAtual)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	mockLogger := &mocks.MockLogger{}
	mockLogger.On("Infof", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Infof", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return()

	processor := NewOrderProcessor(server.URL, mockLogger)

	message := models.Message{
		Topic:     "orders-topic",
		Partition: 0,
		Offset:    1,
		Value:     []byte(`{"ordemDeVenda": "order123", "etapaAtual": "FATURADO"}`),
	}

	ctx := context.Background()
	err := processor.Process(ctx, message)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestOrderProcessor_Process_InvalidJSON(t *testing.T) {
	mockLogger := &mocks.MockLogger{}
	mockLogger.On("Infof", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()

	processor := NewOrderProcessor("http://localhost:8081/api/v1/orders", mockLogger)

	message := models.Message{
		Topic:     "orders-topic",
		Partition: 0,
		Offset:    1,
		Value:     []byte(`invalid json`),
	}

	ctx := context.Background()
	err := processor.Process(ctx, message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
	mockLogger.AssertExpectations(t)
}
