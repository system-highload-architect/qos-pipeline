# 🇬🇧 API Contracts / 🇷🇺 API Контракты

## Общие положения

Взаимодействие с системой `qos-pipeline` осуществляется через REST API, предоставляемый **Gateway**. Все эндпоинты, кроме `/health`, требуют аутентификации OAuth 2.0 (Bearer token).

Форматы данных: JSON.

## Эндпоинты

### Проверка работоспособности

```http
GET /health
```

**Ответ:** `200 OK` с телом `{"status":"ok"}`.

---

### Аутентификация и авторизация

**Регистрация нового клиента:**

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "client_id": "my-service",
  "client_secret": "strong-secret"
}
```

**Ответ:** `201 Created` с телом `{"client_id":"my-service","message":"registered"}`.

**Получение токена:**

```http
POST /api/v1/auth/token
Content-Type: application/json

{
  "client_id": "my-service",
  "client_secret": "strong-secret",
  "grant_type": "client_credentials"
}
```

**Ответ:** `200 OK` с телом `{"access_token":"eyJ...","expires_in":3600}`.

---

### Отправка метрик

Используется **Ingester Service** напрямую (gRPC) или через Gateway (REST).

**REST (через Gateway):**

```http
POST /api/v1/metrics
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "source": "web-server-1",
  "metrics": [
    {
      "name": "http_requests_total",
      "value": 150.0,
      "labels": {"method":"GET","status":"200"},
      "timestamp": "2026-06-27T10:00:00Z"
    },
    {
      "name": "http_request_duration_seconds",
      "value": 0.045,
      "labels": {"method":"GET","status":"200","quantile":"0.95"},
      "timestamp": "2026-06-27T10:00:00Z"
    }
  ]
}
```

**Ответ:** `202 Accepted`.

При превышении лимита — `429 Too Many Requests`.

---

### Получение SLO/SLA отчётов

**Запрос доступности за период:**

```http
GET /api/v1/slo/availability?source=web-server-1&from=2026-06-01T00:00:00Z&to=2026-06-27T23:59:59Z
Authorization: Bearer <access_token>
```

**Ответ:**

```json
{
  "source": "web-server-1",
  "sli": "availability",
  "value": 99.95,
  "target": 99.9,
  "status": "ok",
  "period": "2026-06-01T00:00:00Z/2026-06-27T23:59:59Z"
}
```

**Запрос задержки:**

```http
GET /api/v1/slo/latency?source=web-server-1&from=2026-06-01T00:00:00Z&to=2026-06-27T23:59:59Z&percentile=p95
Authorization: Bearer <access_token>
```

**Ответ:**

```json
{
  "source": "web-server-1",
  "sli": "latency_p95",
  "value": 145.2,
  "target": 200.0,
  "unit": "ms",
  "status": "ok"
}
```

---

### Управление конвейером (администратор)

**Получение статуса стадий:**

```http
GET /api/v1/admin/pipeline/status
Authorization: Bearer <admin_token>
```

**Ответ:**

```json
{
  "stages": [
    {"name":"normalize","workers":4,"backpressure":false,"queue_size":23},
    {"name":"filter","workers":2,"backpressure":false,"queue_size":5},
    {"name":"aggregate","workers":4,"backpressure":true,"queue_size":1950},
    {"name":"write","workers":1,"backpressure":false,"queue_size":10}
  ]
}
```

**Изменение количества воркеров:**

```http
PATCH /api/v1/admin/pipeline/stage/filter
Authorization: Bearer <admin_token>
Content-Type: application/json

{"workers": 4}
```

**Ответ:** `200 OK`.

---

## Модель данных

### Metric

```go
type Metric struct {
    Name      string            `json:"name"`
    Value     float64           `json:"value"`
    Labels    map[string]string `json:"labels,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
}
```

### MetricsBatch

```go
type MetricsBatch struct {
    Source  string   `json:"source"`
    Metrics []Metric `json:"metrics"`
}
```

---

## Обработка ошибок

Все ошибки возвращаются в формате:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "metric name is required"
  }
}
```

Коды HTTP: `400` (некорректный запрос), `401` (неавторизован), `403` (недостаточно прав), `429` (превышен лимит), `500` (внутренняя ошибка).

---

## Версионирование

API версионируется через префикс `/api/v1/`. Обратная совместимость гарантируется в рамках мажорной версии. Новая функциональность может добавляться без изменения существующих эндпоинтов.