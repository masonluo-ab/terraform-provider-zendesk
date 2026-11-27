package client

import (
	"context"
	"net/http"
	"sync"
)

// TokenExchanger returns an access token for the OAuth client a transport authenticates as.
type TokenExchanger func(ctx context.Context) (string, error)

// OAuthTransport sets a bearer token on every request it sends, exchanging the OAuth client's credentials for
// the token on the first request rather than when the provider is configured. A run that sends no request, such
// as a plan with refresh off, never contacts the token endpoint, so it needs no valid credentials.
//
// The token is exchanged once and reused; it is not renewed when it expires.
type OAuthTransport struct {
	base     http.RoundTripper
	exchange TokenExchanger

	mu    sync.Mutex
	token string
}

// NewOAuthTransport sends through base, or http.DefaultTransport when base is nil. exchange must not send
// through the returned transport.
func NewOAuthTransport(base http.RoundTripper, exchange TokenExchanger) *OAuthTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &OAuthTransport{base: base, exchange: exchange}
}

// RoundTrip sends req with the token in its Authorization header, replacing whatever the caller set there.
func (t *OAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.accessToken(req.Context())
	if err != nil {
		return nil, err
	}

	authorized := req.Clone(req.Context())
	authorized.Header.Set("Authorization", "Bearer "+token)
	return t.base.RoundTrip(authorized)
}

// accessToken holds the lock across the exchange, so concurrent first requests share one exchange. A failed
// exchange is not cached: the next request tries again.
func (t *OAuthTransport) accessToken(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.token == "" {
		token, err := t.exchange(ctx)
		if err != nil {
			return "", err
		}
		t.token = token
	}
	return t.token, nil
}
