package embrace

import "github.com/hossein-nas/analytics_aggregator/internal/project/collector"

type Config struct {
	collector.BaseCollector `bson:",inline"`
	AppID                   string `bson:"app_id"`
	APIKey                  string `bson:"api_key"`
	Platform                string `bson:"platform"` // ios/android
	Host                    string `bson:"host"`
}
