package mongo_schema

type RunnerGroupSchema struct {
	GroupName string   `json:"group_name"`
	GroupCode string   `json:"group_code"`
	CreatedAt int64    `json:"created_at"`
	Members   []string `json:"members"`
	Owner     string   `json:"owner"`
}
