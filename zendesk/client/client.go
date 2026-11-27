package client

import (
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/nukosuke/go-zendesk/zendesk"
)

type (
	Client struct {
		// use struct embedding for extension
		zendesk.Client

		// go-zendesk keeps its base URL, credential and HTTP client private and has no PATCH helper, so the
		// requests it cannot make are sent with these copies.
		baseURL    string
		credential zendesk.Credential
		httpClient *http.Client
	}
)

// NewClient wraps zd. baseURL is the instance's API root (https://<account>.zendesk.com/api/v2), and credential
// and httpClient are the ones zd sends its own requests with.
func NewClient(zd zendesk.Client, baseURL string, credential zendesk.Credential, httpClient *http.Client) *Client {
	return &Client{
		Client:     zd,
		baseURL:    baseURL,
		credential: credential,
		httpClient: httpClient,
	}
}

// addOptions build query string
func addOptions(s string, opts interface{}) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
