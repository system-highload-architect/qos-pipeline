# 🇬🇧 Launch Guide / 🇷🇺 Инструкция по запуску

This document describes how to run the system: from quick Docker deployment to local development of individual services.
Этот документ описывает способы запуска системы: от быстрого развёртывания в Docker до локальной разработки отдельных сервисов.

---

## 🇬🇧 Requirements / 🇷🇺 Требования

- **Docker** and **Docker Compose** (v2 or newer)
- **Node.js 22+** (only for running the dashboard locally outside a container)
- Free ports: `8080`, `9001–9003`, `5432`, `9092`, `2181`

---

## 🇬🇧 Preparing the Go workspace / 🇷🇺 Подготовка Go workspace

**This step is mandatory before any other operation** (both Docker and local development).  
**Этот шаг обязателен перед всеми остальными** (как для Docker, так и для локальной разработки).

The project uses a **Go workspace** to manage multiple interdependent modules without `replace` directives.  
Проект использует **Go workspace** для управления несколькими взаимозависимыми модулями без директив `replace`.

Run the following script **from the project root** – it will initialise the workspace, download all dependencies, and build every service.  
Выполните следующий скрипт **из корня проекта** – он инициализирует workspace, скачает все зависимости и соберёт все сервисы.

```bash
# Инициализируем workspace, если его нет
if [ ! -f go.work ]; then
  go work init
  go work use -r .
fi

# Синхронизируем и подтягиваем зависимости
go work sync
find . -name "go.mod" -execdir go mod tidy \;
go build ./...
```

**What this script does / Что делает скрипт:**
1. Creates `go.work` (if missing) and automatically adds all modules.  
   Создаёт `go.work` (если он отсутствует) и автоматически добавляет в него все модули.
2. Synchronises dependencies between modules.  
   Синхронизирует зависимости между модулями.
3. Runs `go mod tidy` in every module to fetch missing packages.  
   Выполняет `go mod tidy` в каждом модуле, чтобы подтянуть недостающие пакеты.
4. Builds all services.  
   Собирает все сервисы.

Now you are ready to start the platform – either via Docker or locally.  
Теперь вы готовы запустить платформу – либо через Docker, либо локально.

---

## 🇬🇧 Quick Start (Production‑like environment) / 🇷🇺 Быстрый старт (Production‑like окружение)

1. Clone the repository and navigate to its root:
   ```bash
   git clone https://github.com/system-highload-architect/qos-pipeline
   cd qos-pipeline
   ```

2. **Complete the workspace preparation described above.**  
   **Выполните подготовку workspace, описанную выше.**

3. Start all services and databases with a single command:
   ```bash
   docker compose up --build -d
   ```
   The first build may take a few minutes. Subsequent runs (`docker compose up -d`) will be instant.

4. Open [http://localhost:8080](http://localhost:8080) in your browser – this is the dashboard.  
   **Note:** In the current demo version, the dashboard is not yet available. Use the GraphQL playground instead: [http://localhost:8080/graphql](http://localhost:8080/graphql).

5. **Done!** You can send test metrics and query GraphQL.

---

## 🇬🇧 Configuration / 🇷🇺 Конфигурация

All services are configured via environment variables with the `RTB_` prefix (to match our `pkg/config`).  
Below are the main parameters you may need in production:

| Variable | Description | Default value in Docker |
|----------|-------------|-------------------------|
| `SERVER_PORT` | HTTP port for Gateway / gRPC port for other services | `8080` (Gateway), `9001` (Ingester), `9002` (Aggregator) |
| `KAFKA_BROKERS` | Kafka broker list | `kafka:29092` |
| `KAFKA_TOPIC` | Topic for raw metrics | `raw_metrics` |
| `DATABASE_DSN` | PostgreSQL DSN (Aggregator) | `postgres://rtb:rtbpass@postgres:5432/rtb?sslmode=disable` |
| `GRPC_INGESTER` | Ingester gRPC address (for Gateway) | `ingester:9001` |
| `GRPC_AGGREGATOR` | Aggregator HTTP address (for Gateway GraphQL proxy) | `http://aggregator:9002` |

Full variable lists are available in `docker-compose.yml` and `services/*/configs/dev.yaml`.

---

## 🇬🇧 Local development (without Docker for Go) / 🇷🇺 Локальная разработка (без Docker для Go)

If you want to run services natively (Go) while databases run in Docker:

1. Complete the **workspace preparation** described at the beginning.  
   Выполните **подготовку workspace**, описанную в начале.

2. Start the infrastructure containers:
   ```bash
   docker compose up -d postgres kafka
   ```

3. In separate terminals, launch the services:
   ```bash
   # Ingester
   cd services/ingester && go run ./cmd
   # Pipeline
   cd services/pipeline && go run ./cmd
   # Aggregator
   cd services/aggregator && go run ./cmd
   # Gateway
   cd services/gateway && go run ./cmd
   ```

   Services read configuration from `configs/dev.yaml`, where database addresses are `localhost`. Ensure the corresponding ports are exposed in `docker-compose.yml`.

4. The Gateway will serve the GraphQL playground at `http://localhost:8080/graphql`.

---

## 🇬🇧 Testing / 🇷🇺 Тестирование

Send a test metric to the Gateway:

```bash
curl -X POST http://localhost:8080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{"source":"web-1","metrics":[{"name":"cpu","value":0.5,"timestamp":1719500000}]}'
```

Expected response: `{"status":"accepted"}`.

Then query GraphQL for SLO status:

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ sloStatus(source: \"web-1\") { source available p95 p99 status } }"}'
```

You should see a JSON response with demo data.

---

## 🇬🇧 Troubleshooting / 🇷🇺 Устранение неполадок

### 1. Containers do not start or crash with an error
- Check logs: `docker compose logs <service>`
- Ensure ports are not occupied by other processes.
- Make sure Docker Desktop is running in Linux container mode.

### 2. “connection refused” between services
- Ensure all services use correct Docker network hostnames (e.g., `kafka:29092` instead of `localhost:29092`). This is configured in `docker-compose.yml` via environment variables.
- Verify that all containers are on the same network (Docker Compose creates one automatically).

### 3. Kafka errors (“Unknown Topic Or Partition”)
- The topic `raw_metrics` is created automatically by Kafka. Wait a few seconds after startup and retry.
- If the problem persists, check that Kafka and Zookeeper are healthy (`docker compose ps`).

### 4. GraphQL returns “connection refused” in Gateway logs
- Ensure the Aggregator is running and Gateway uses the correct address (`aggregator:9002`). Check Gateway's `cmd/main.go` for `WithAggregatorURL`.

---

## 🇬🇧 Stopping the system / 🇷🇺 Остановка системы

```bash
docker compose down
```

To also remove all data (volumes), add the `-v` flag:
```bash
docker compose down -v
```

---

## 🇬🇧 Additional information / 🇷🇺 Дополнительная информация

- Architecture and service descriptions: [docs/specification/](docs/specification/)
- Feature specifications: [docs/srs/](docs/srs/)
- Configuration and environment variables: `docker-compose.yml` and `services/*/configs/`