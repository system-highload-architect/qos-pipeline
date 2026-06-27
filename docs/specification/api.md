# 🇬🇧 API Contracts / 🇷🇺 API Контракты

## Общие положения
Взаимодействие с платформой осуществляется через **Gateway** по REST и GraphQL.

## REST API

### Приём метрик
```http
POST /api/v1/metrics
Content-Type: application/json

{
  "source": "web-server-1",
  "metrics": [
    {
      "name": "http_requests_total",
      "value": 150.0,
      "labels": {"method": "GET", "status": "200"},
      "timestamp": 1719500000
    }
  ]
}
```
**Ответ:** `202 Accepted` с телом `{"status":"accepted"}`.

## GraphQL API

Эндпоинт: `POST /graphql`

### Запрос агрегированных метрик
```graphql
{
  metrics(source: "web-1", name: "http_requests_total", limit: 10) {
    source
    name
    sum
    count
    p95
    p99
    min
    max
    createdAt
  }
}
```

### Запрос статуса SLO
```graphql
{
  sloStatus(source: "web-1") {
    source
    available
    p95
    p99
    status
  }
}
```

### Запрос истории SLO
```graphql
{
  sloHistory(source: "web-1", from: "2026-06-01T00:00:00Z", to: "2026-06-27T23:59:59Z") {
    date
    available
    p95
    p99
    status
  }
}
```

## Модели данных

### Metric
| Поле | Тип | Описание |
|------|-----|----------|
| `id` | Int | Идентификатор записи |
| `source` | String | Источник метрики |
| `name` | String | Название метрики |
| `sum` | Float | Сумма значений |
| `count` | Int | Количество измерений |
| `min` | Float | Минимальное значение |
| `max` | Float | Максимальное значение |
| `p50` | Float | 50-й процентиль |
| `p95` | Float | 95-й процентиль |
| `p99` | Float | 99-й процентиль |
| `createdAt` | DateTime | Время создания записи |

### SLOStatus
| Поле | Тип | Описание |
|------|-----|----------|
| `source` | String | Источник |
| `available` | Float | Доступность в процентах |
| `p95` | Float | 95-й процентиль задержки (мс) |
| `p99` | Float | 99-й процентиль задержки (мс) |
| `status` | String | Статус: ok, warning, breach |

### SLOHistory
| Поле | Тип | Описание |
|------|-----|----------|
| `date` | DateTime | Дата записи |
| `available` | Float | Доступность в процентах |
| `p95` | Float | 95-й процентиль задержки (мс) |
| `p99` | Float | 99-й процентиль задержки (мс) |
| `status` | String | Статус: ok, warning, breach |