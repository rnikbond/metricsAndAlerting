package storage

import (
	"metrics-and-alerting/pkg/metric"
)

type Repository interface {
	Upsert(metric metric.Metric) error
	UpsertSlice(metrics []metric.Metric) error

	Get(metric metric.Metric) (metric.Metric, error)
	GetSlice() ([]metric.Metric, error)

	Delete(metric metric.Metric) error

	CheckHealth() bool
	Close() error
}
