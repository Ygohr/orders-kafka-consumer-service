package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/Ygohr/orders-kafka-consumer-service/internal/consumer/models"
)

type DLQService struct {
	producer sarama.SyncProducer
	topic    string
}

func NewDLQService(brokers []string, topic string, username, password string) (*DLQService, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 10 * time.Second

	if username != "" && password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		config.Net.SASL.User = username
		config.Net.SASL.Password = password
	}

	config.Version = sarama.V2_8_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
	}

	return &DLQService{
		producer: producer,
		topic:    topic,
	}, nil
}

func (d *DLQService) SendToDLQ(ctx context.Context, originalMsg models.Message, err error, retryCount int) error {
	dlqMessage := models.NewDLQMessage(originalMsg, err, retryCount)

	jsonData, err := dlqMessage.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: d.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%s-%d-%d", originalMsg.Topic, originalMsg.Partition, originalMsg.Offset)),
		Value: sarama.ByteEncoder(jsonData),
		Headers: []sarama.RecordHeader{
			{Key: []byte("original_topic"), Value: []byte(originalMsg.Topic)},
			{Key: []byte("error_type"), Value: []byte("processing_error")},
			{Key: []byte("retry_count"), Value: []byte(fmt.Sprintf("%d", retryCount))},
		},
	}

	partition, offset, err := d.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	log.Printf("Message sent to DLQ topic: %s, partition: %d, offset: %d, original topic: %s, error: %v",
		d.topic, partition, offset, originalMsg.Topic, err)

	return nil
}

func (d *DLQService) Close() error {
	return d.producer.Close()
}
