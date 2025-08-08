# Orders Kafka Consumer Service

Kafka message consumption service developed in Go for asynchronous processing of sales orders.

## Architecture

### Clean Architecture + SOLID

- **Separation of responsibilities**: Consumer, Processor, Validator, DLQ
- **Well-defined interfaces**: Dependency Injection and dependency inversion
- **Extensibility**: New processors and validators without modifying existing code

### Structure

```
internal/
├── app/          # Application orchestration
├── config/       # Configuration management
├── consumer/     # Kafka consumption
├── service/      # Order processing
└── logger/       # Structured logging
```

## Technologies

- **Go 1.24.4** + **Apache Kafka** + **IBM Sarama**
- **Zap** (logger) + **Viper** (config) + **Testify** (tests)
- **Docker Compose** (local infrastructure)

## Processing Flow

```
Kafka → Validation → Processing → HTTP PATCH → DLQ (if error)
```

## Message Format

The service processes order messages in JSON format. See `payload_example.json` for the expected structure:

```json
{
  "ordemDeVenda": "order12345",
  "etapaAtual": "FATURADO"
}
```

## Local Execution

### Prerequisites

- Go 1.24.4+
- Docker + Docker Compose
- Git

### Environment Scripts

The project includes scripts to facilitate local environment management:

```bash
# Start local environment
./start-local.sh

# Stop local environment
./stop-local.sh
```

### Execution Steps

#### 1. Clone and Setup

```bash
git clone <repository-url>
cd orders-kafka-consumer-service
go mod download
```

#### 2. Environment Configuration

```bash
# Copy environment file (if not exists)
cp .env.example .env

# Edit configuration as needed
nano .env
```

#### 3. Start Local Environment

```bash
./start-local.sh
```

This script will:

- Start Kafka, Zookeeper, and Kafka UI
- Wait for services to be ready
- Display connection information

#### 4. Start Mock Service

```bash
cd mock-service
go run main.go
```

#### 5. Start Consumer Service

```bash
# In a new terminal
cd orders-kafka-consumer-service
go run main.go
```

#### 6. Verification

- **Kafka UI**: http://localhost:8080
- **Mock Service**: http://localhost:8081

## Tests

### Execution

```bash
# Run all tests
go test ./... -v -cover
```

### Test Coverage

| Package    | Coverage | Status | Tests |
| ---------- | -------- | ------ | ----- |
| `consumer` | 100%     | ✅     | 6     |
| `service`  | 50%      | ⚠️     | 3     |
| `config`   | 0%       | ❌     | 0     |
| `app`      | 0%       | ❌     | 0     |
| `logger`   | 0%       | ❌     | 0     |

**Total**: 9 tests, 100% success, ~0.8s execution

**Average coverage**: 30%
