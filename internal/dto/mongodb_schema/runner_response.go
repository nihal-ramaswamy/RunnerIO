package mongo_schema

type RunnerResponse struct {
	Polygons  []RunnerPolygonSchema   `json:"polygons"`
	LiveLines []RunnerLiveLinesData `json:"liveLines"`
}
