package dtoschema

import "time"

type CoordinateStruct struct {
	X    float64   `json:"x"`
	Y    float64   `json:"y"`
	Time time.Time `json:"time"`
}

type PolygonData struct {
	Coords       []CoordinateStruct `json:"coordinates"`
	Runner       string             `json:"runner"`
	Time         time.Time          `json:"time"`
	InsertedTime time.Time          `json:"inserted_time"`
}

type userData struct {
	UserSub  string        `json:"user_sub"`
	Polygons []PolygonData `json:"polygons"`
	LiveLine PolygonData   `json:"live_line"`
}

type RunnerEatAreaSchema struct {
	GroupCode string `json:"group_code"`
	UserData  []userData
}
