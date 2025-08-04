package mongo_schema

import "time"

type RunnerLiveLinesData struct {
	X            float64   `json:"x" bson:"x"`
	Y            float64   `json:"y" bson:"y"`
	Time         time.Time `json:"time" bson:"time"`
	Sender       string    `json:"sender" bson:"sender"`
	GroupCode    string    `json:"group_code" bson:"group_code"`
	InsertedTime int       `json:"inserted_time" bson:"inserted_time"`
}

func NewRunnerLiveLinesData(x, y float64, time time.Time, sender, groupCode string) RunnerLiveLinesData {
	return RunnerLiveLinesData{
		X:         x,
		Y:         y,
		Time:      time,
		Sender:    sender,
		GroupCode: groupCode,
	}
}

func (liveLinesData RunnerLiveLinesData) GetInsertedTime() int {
	return liveLinesData.InsertedTime
}

func ToPolygonData(runnerLiveLinesSchema *[]RunnerLiveLinesData) RunnerPolygonSchema {
	maxTime := (*runnerLiveLinesSchema)[len(*runnerLiveLinesSchema)-1].Time
	coords := []CoordinateStruct{}
	for _, liveLine := range *runnerLiveLinesSchema {
		coords = append(coords, CoordinateStruct{
			X:    liveLine.X,
			Y:    liveLine.Y,
			Time: liveLine.Time,
		})
		if liveLine.Time.After(maxTime) {
			maxTime = liveLine.Time
		}
	}

	return RunnerPolygonSchema{
		Time:   maxTime,
		Coords: coords,
		Runner: (*runnerLiveLinesSchema)[0].Sender,
	}
}
