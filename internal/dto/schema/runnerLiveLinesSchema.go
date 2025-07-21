package dtoschema

import "time"

type LiveLinesData struct {
	X         float64   `json:"x" bson:"x"`
	Y         float64   `json:"y" bson:"y"`
	Time      time.Time `json:"time" bson:"time"`
	Sender    string    `json:"sender" bson:"sender"`
	GroupCode string    `json:"group_code" bson:"group_code"`
}

func ToPolygonData(runnerLiveLinesSchema *[]LiveLinesData) *PolygonData {
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

	return &PolygonData{
		Time:   maxTime,
		Coords: coords,
		Runner: (*runnerLiveLinesSchema)[0].Sender,
	}
}

type TestGenerics interface {
	LiveLinesData | PolygonData
}
