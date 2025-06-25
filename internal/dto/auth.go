package dto

type UserData struct {
	Nickname  string `json:"nickname"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	UpdatedAt string `json:"updated_at"`
	Sub       string `json:"sub"`
}

type MgmtPostRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
	GrantType    string `json:"grant_type"`
}

type MgmtPostResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func NewMgmtPostRequest(clientId, clientSecret, audience string) *MgmtPostRequest {
	return &MgmtPostRequest{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Audience:     audience,
		GrantType:    "client_credentials",
	}
}
