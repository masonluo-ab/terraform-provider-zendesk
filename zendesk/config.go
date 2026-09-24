package zendesk

import "fmt"

// Config is configuration struct for Zendesk credentials
type Config struct {
	Account           string
	Email             string
	Token             string
	OAuthClientID     string
	OAuthClientSecret string
	OAuthScope        string
}

// usesOAuth reports whether the configuration authenticates with an OAuth client rather than an API token,
// and rejects anything but exactly one complete pair: email with token, or OAuth client id with secret.
func (c Config) usesOAuth() (bool, error) {
	apiToken := c.Email != "" || c.Token != ""
	oauth := c.OAuthClientID != "" || c.OAuthClientSecret != ""

	switch {
	case apiToken && oauth:
		return false, fmt.Errorf("configure either email and token, or oauth_client_id and oauth_client_secret, not both")
	case oauth:
		if c.OAuthClientID == "" || c.OAuthClientSecret == "" {
			return false, fmt.Errorf("oauth_client_id and oauth_client_secret must be set together")
		}
		return true, nil
	case apiToken:
		if c.Email == "" || c.Token == "" {
			return false, fmt.Errorf("email and token must be set together")
		}
		return false, nil
	default:
		return false, fmt.Errorf("no credentials: configure email and token, or oauth_client_id and oauth_client_secret")
	}
}
