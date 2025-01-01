package clarity

import "github.com/hossein-nas/analytics_aggregator/internal/project/collector"

type Config struct {
	collector.BaseCollector `bson:",inline"`
	ProjectID               string `json:"project_id" bson:"project_id"`
	APIKey                  string `json:"api_key" bson:"api_key"`
}
