package mocks

import (
	"context"

	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer"
	"github.com/stretchr/testify/mock"
)

type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockConsumer) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockConsumer) Subscribe(topic string, handler consumer.MessageHandler) error {
	args := m.Called(topic, handler)
	return args.Error(0)
}

func (m *MockConsumer) AddValidator(topic string, validator consumer.MessageValidator) error {
	args := m.Called(topic, validator)
	return args.Error(0)
}

func (m *MockConsumer) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}
