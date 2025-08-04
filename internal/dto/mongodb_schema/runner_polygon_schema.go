package mongo_schema

import "time"

type RunnerPolygonSchema struct {
	Coords       []CoordinateStruct `json:"coordinates" bson:"coordinates"`
	Runner       string             `json:"runner" bson:"runner"`
	Time         time.Time          `json:"time" bson:"time"`
	InsertedTime int                `json:"inserted_time" bson:"inserted_time"`
	GroupCode    string             `json:"group_code" bson:"group_code"`
}

func NewRunnerPolygonSchema(runner, groupCode string, coords []CoordinateStruct, time time.Time) RunnerPolygonSchema {
	return RunnerPolygonSchema{
		Coords:    coords,
		Runner:    runner,
		GroupCode: groupCode,
		Time:      time,
	}
}

func (polygonData RunnerPolygonSchema) GetInsertedTime() int {
	return polygonData.InsertedTime
}
