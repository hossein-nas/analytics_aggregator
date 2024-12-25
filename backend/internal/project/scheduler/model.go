package scheduler

type Stat struct {
	ID      string      `json:"_id" bson:"_id"`
	Type    string      `json:"type" validate:"required,oneof=clarity sentry" bson:"type"`
	LastRun string      `json:"last_run" bson:"last_run" validate:"required"`
	Data    interface{} `json:"data" bson:"data" validate="required"`
}

type UpdateStat struct {
	Data interface{} `json:"data" bson:"data" validate="required"`
}
