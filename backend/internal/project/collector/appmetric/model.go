package appmetric

import model "github.com/hossein-nas/analytics_aggregator/internal/project/models"

type Config struct {
	model.BaseCollector `bson:",inline"`
	ApplicationID       string   `bson:"application_id"`
	APIKey              string   `bson:"api_key"`
	Metrics             []string `bson:"metrics"` // List of metrics to collect
	Host                string   `bson:"host"`
}
