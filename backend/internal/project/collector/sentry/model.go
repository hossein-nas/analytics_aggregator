package sentry

import "github.com/hossein-nas/analytics_aggregator/internal/project/collector"

type Config struct {
	collector.BaseCollector `bson:",inline"`
	OrganizationSlug        string `json:"org_slug" bson:"org_slug"`
	ProjectSlug             string `json:"project_slug" bson:"project_slug"`
	AuthToken               string `json:"auth_token" bson:"auth_token"`
	Host                    string `json:"host,omitempty" bson:"host,omitempty"` // Optional for self-hosted
}
