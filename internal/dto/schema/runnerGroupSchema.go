package dtoschema

type RunnerGroupSchema struct {
	Group     string   `json:"group"`
	Code      string   `json:"code"`
	CreatedAt int64    `json:"created_at"`
	Members   []string `json:"members"`
	Owner     string   `json:"owner"`
}
