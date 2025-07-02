package dtoschema

import "time"

type RunnerAuditSchema struct {
	X         float64   `json:"x" bson:"x"`
	Y         float64   `json:"y" bson:"y"`
	Time      time.Time `json:"time" bson:"time"`
	Sender    string    `json:"sender" bson:"sender"`
	GroupCode string    `json:"group_code" bson:"group_code"`
}

type RunnerAuditSchemaRequest struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Time      int64   `json:"time"`
	GroupCode string  `json:"group_code"`
}

func (r *RunnerAuditSchemaRequest) ToRunnerAuditSchema(sender string) *RunnerAuditSchema {
	return &RunnerAuditSchema{
		X:         r.X,
		Y:         r.Y,
		Time:      time.UnixMilli(r.Time),
		Sender:    sender,
		GroupCode: r.GroupCode,
	}
}
