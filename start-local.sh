#!/bin/bash

echo "Starting local environment..."

if ! command -v docker &> /dev/null; then
    echo "Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "Docker Compose is not installed"
    exit 1
fi

echo "Starting services..."
docker-compose up -d

echo "Waiting for services to be ready..."
sleep 30

echo "Environment started successfully!"
echo ""
echo "Kafka UI: http://localhost:8080"
echo "Kafka: localhost:9092"
echo ""
echo "To test:"
echo "1. cp env.example .env"
echo "2. go run main.go"
echo ""
echo "To stop: ./stop-local.sh"
