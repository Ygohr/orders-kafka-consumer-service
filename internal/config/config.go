package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	KafkaBootstrapServers string `mapstructure:"KAFKA_BOOTSTRAP_SERVERS"`
	KafkaUsername         string `mapstructure:"KAFKA_USERNAME"`
	KafkaPassword         string `mapstructure:"KAFKA_PASSWORD"`
	KafkaSaslMechanism    string `mapstructure:"KAFKA_SASL_MECHANISM"`
	KafkaSecurityProtocol string `mapstructure:"KAFKA_SECURITY_PROTOCOL"`
	KafkaGroupId          string `mapstructure:"KAFKA_GROUP_ID"`
	KafkaTopic            string `mapstructure:"KAFKA_TOPIC"`
	KafkaTopicDlq         string `mapstructure:"KAFKA_TOPIC_DLQ"`
	TargetServiceUrl      string `mapstructure:"TARGET_SERVICE_URL"`
	LogLevel              string `mapstructure:"LOG_LEVEL"`
	LogDevelopment        bool   `mapstructure:"LOG_DEVELOPMENT"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	viper.SetDefault("KAFKA_USERNAME", "exemplo")
	viper.SetDefault("KAFKA_PASSWORD", "exemplo")
	viper.SetDefault("KAFKA_SASL_MECHANISM", "SCRAM-SHA-256")
	viper.SetDefault("KAFKA_SECURITY_PROTOCOL", "SASL_PLAINTEXT")
	viper.SetDefault("KAFKA_GROUP_ID", "seu-grupo")
	viper.SetDefault("KAFKA_TOPIC", "seu-topico")
	viper.SetDefault("KAFKA_TOPIC_DLQ", "seu-topico-dlq")
	viper.SetDefault("TARGET_SERVICE_URL", "http://localhost:8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_DEVELOPMENT", true)

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading .env file: %w", err)
		}
	}

	viper.AutomaticEnv()

	var config *Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.New("failed to unmarshal config")
	}

	return config, nil
}
