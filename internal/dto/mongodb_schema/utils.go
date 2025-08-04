package mongo_schema

import "time"

type CoordinateStruct struct {
	X    float64   `json:"x"`
	Y    float64   `json:"y"`
	Time time.Time `json:"time"`
}

type MongoDataInterface interface {
	RunnerPolygonSchema | RunnerLiveLinesData
	GetInsertedTime() int
}
