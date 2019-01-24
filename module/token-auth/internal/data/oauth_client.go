package data

type OAuthClient struct {
	ID          string `json:"client_id"`
	Secret      string `json:"client_secret"`
	RedirectURI string `json:"redirect_uri`
	Active      bool   `json:"active"`
}

type OauthTokenInto struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}
