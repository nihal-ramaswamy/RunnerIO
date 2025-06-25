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

type sources struct {
	SourceId   string `json:"source_id"`
	SourceName string `json:"source_name"`
	SourceType string `json:"source_type"`
}

type UserPermission struct {
	PermissionName     string    `json:"permission_name"`
	Description        string    `json:"description"`
	ResourceServerName string    `json:"resource_server_name"`
	ResourceServerID   string    `json:"resource_server_identifier"`
	Sources            []sources `json:"sources"`
}

func NewMgmtPostRequest(clientId, clientSecret, audience string) *MgmtPostRequest {
	return &MgmtPostRequest{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Audience:     audience,
		GrantType:    "client_credentials",
	}
}
