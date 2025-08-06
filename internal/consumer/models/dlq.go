package models

import (
	"encoding/json"
	"time"
)

type DLQMessage struct {
	OriginalMessage Message   `json:"original_message"`
	Error           string    `json:"error"`
	FailedAt        time.Time `json:"failed_at"`
	RetryCount      int       `json:"retry_count"`
	Topic           string    `json:"topic"`
	Partition       int32     `json:"partition"`
	Offset          int64     `json:"offset"`
}

func NewDLQMessage(originalMsg Message, err error, retryCount int) DLQMessage {
	return DLQMessage{
		OriginalMessage: originalMsg,
		Error:           err.Error(),
		FailedAt:        time.Now(),
		RetryCount:      retryCount,
		Topic:           originalMsg.Topic,
		Partition:       originalMsg.Partition,
		Offset:          originalMsg.Offset,
	}
}

func (d DLQMessage) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}
