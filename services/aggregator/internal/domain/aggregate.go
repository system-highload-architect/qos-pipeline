package domain

import "time"

// AggregatedMetric — метрика, полученная после агрегации.
type AggregatedMetric struct {
	Source    string
	Name      string
	Sum       float64
	Count     int64
	Min       float64
	Max       float64
	P50       float64
	P95       float64
	P99       float64
	Window    time.Duration
	CreatedAt time.Time
}
