package clarity

import model "github.com/hossein-nas/analytics_aggregator/internal/project/models"

type Config struct {
	model.BaseCollector `bson:",inline"`
	ProjectID           string `bson:"project_id"`
	APIKey              string `bson:"api_key"`
	Host                string `bson:"host"`
}
