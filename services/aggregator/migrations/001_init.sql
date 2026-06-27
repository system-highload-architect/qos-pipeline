-- Таблица для хранения входящих агрегатов
CREATE TABLE IF NOT EXISTS metrics (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sum DOUBLE PRECISION NOT NULL DEFAULT 0,
    count BIGINT NOT NULL DEFAULT 0,
    min DOUBLE PRECISION,
    max DOUBLE PRECISION,
    p50 DOUBLE PRECISION,
    p95 DOUBLE PRECISION,
    p99 DOUBLE PRECISION,
    window_seconds BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для быстрых запросов
CREATE INDEX IF NOT EXISTS idx_metrics_source ON metrics(source);
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);
CREATE INDEX IF NOT EXISTS idx_metrics_created ON metrics(created_at);

-- Определения SLO
CREATE TABLE IF NOT EXISTS slo_definitions (
    id SERIAL PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    target_value DOUBLE PRECISION NOT NULL,
    operator VARCHAR(10) NOT NULL DEFAULT 'gte', -- gte, lte
    window_days INT NOT NULL DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Результаты расчётов SLO
CREATE TABLE IF NOT EXISTS slo_results (
    id BIGSERIAL PRIMARY KEY,
    slo_id INT REFERENCES slo_definitions(id),
    actual_value DOUBLE PRECISION NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    compliant BOOLEAN NOT NULL,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);