package sentry

import model "github.com/hossein-nas/analytics_aggregator/internal/project/models"

type Config struct {
	model.BaseCollector `bson:",inline"`
	OrganizationSlug    string `bson:"org_slug"`
	ProjectSlug         string `bson:"project_slug"`
	AuthToken           string `bson:"auth_token"`
	Host                string `bson:"host,omitempty"` // Optional for self-hosted
}
