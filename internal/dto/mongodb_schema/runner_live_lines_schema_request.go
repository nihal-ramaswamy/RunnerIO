package mongo_schema

import "time"

type RunnerLiveLinesSchemaRequest struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Time      int64   `json:"time"`
	GroupCode string  `json:"group_code"`
}

func (r *RunnerLiveLinesSchemaRequest) ToRunnerLiveLinesSchema(sender string) *RunnerLiveLinesData {
	return &RunnerLiveLinesData{
		X:         r.X,
		Y:         r.Y,
		Time:      time.UnixMilli(r.Time),
		Sender:    sender,
		GroupCode: r.GroupCode,
	}
}
