package embrace

import "github.com/hossein-nas/analytics_aggregator/internal/project/collector"

type Config struct {
	collector.BaseCollector `bson:",inline"`
	AppID                   string `json:"app_id" bson:"app_id"`
	APIKey                  string `json:"api_key" bson:"api_key"`
	Platform                string `json:"platform" bson:"platform"` // ios/android
	Host                    string `json:"host" bson:"host"`
	AppSecret               string `json:"app_secret" bson:"app_secret"`
}
