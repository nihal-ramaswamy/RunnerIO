package dtoschema

type RunnerResponse struct {
	Polygons  []PolygonData   `json:"polygons"`
	LiveLines []LiveLinesData `json:"liveLines"`
}
