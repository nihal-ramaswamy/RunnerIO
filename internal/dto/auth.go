package dto

type UserData struct {
	Nickname  string `json:"nickname"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	UpdatedAt string `json:"updated_at"`
	Sub       string `json:"sub"`
}
