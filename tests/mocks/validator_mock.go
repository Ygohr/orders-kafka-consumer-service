package mocks

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
)

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(msg *sarama.ConsumerMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}
