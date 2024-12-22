package project

import (
	"time"
)

type ServiceType string

const (
	ServiceSentry     ServiceType = "sentry"
	ServiceClarity    ServiceType = "clarity"
	ServiceEmbrace    ServiceType = "embrace"
	ServiceAppMetrica ServiceType = "appmetrica"
)

type ServiceConfig struct {
	Type        ServiceType `json:"type" bson:"type"`
	ServerURL   string      `json:"server_url" bson:"server_url"`
	OrgID       string      `json:"org_id" bson:"org_id"`
	ProjectName string      `json:"project_name" bson:"project_name"`
	AuthToken   string      `json:"auth_token" bson:"auth_token"`
	// Additional fields can be added based on service requirements
}

type Project struct {
	ID            string          `json:"id" bson:"_id"`
	Name          string          `json:"name" bson:"name"`
	Key           string          `json:"key" bson:"key"`
	Services      []ServiceConfig `json:"services" bson:"services"`
	CreatedBy     string          `json:"created_by" bson:"created_by"`
	CreatedAt     time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" bson:"updated_at"`
	LastCollected time.Time       `json:"last_collected" bson:"last_collected"`
	Active        bool            `json:"active" bson:"active"`
}
