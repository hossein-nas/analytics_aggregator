package appmetric

import "github.com/hossein-nas/analytics_aggregator/internal/project/collector"

type Config struct {
	collector.BaseCollector `bson:",inline"`
	ApplicationID           string   `bson:"application_id"`
	APIKey                  string   `bson:"api_key"`
	Metrics                 []string `bson:"metrics"` // List of metrics to collect
	Host                    string   `bson:"host"`
}
