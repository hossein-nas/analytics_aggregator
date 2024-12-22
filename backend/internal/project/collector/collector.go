package collector

import "context"

type MetricsCollector interface {
	Validate() error
	Collect(ctx context.Context) error
	GetMetrics() (map[string]interface{}, error)
}
