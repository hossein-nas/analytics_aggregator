package embrace

import model "github.com/hossein-nas/analytics_aggregator/internal/project/models"

type Config struct {
	model.BaseCollector `bson:",inline"`
	AppID               string `bson:"app_id"`
	APIKey              string `bson:"api_key"`
	Platform            string `bson:"platform"` // ios/android
	Host                string `bson:"host"`
}
