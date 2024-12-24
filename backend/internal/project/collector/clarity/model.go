package clarity

import "github.com/hossein-nas/analytics_aggregator/internal/project/collector"

type Config struct {
	collector.BaseCollector `bson:",inline"`
	ProjectID               string `bson:"project_id"`
	APIKey                  string `bson:"api_key"`
	Host                    string `bson:"host"`
}
