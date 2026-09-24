package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type clientCredentialsRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
}

type clientCredentialsResponse struct {
	AccessToken string `json:"access_token"`
}

// ExchangeClientCredentials exchanges a confidential OAuth client's identifier and secret for an access token
// with the client_credentials grant. baseURL is the instance root, e.g. https://example.zendesk.com: the token
// endpoint sits outside /api/v2. An empty scope is left out of the request.
//
// The token acts as the user who owns the OAuth client, and expires after about 30 minutes; it is not renewed.
// ref: https://developer.zendesk.com/api-reference/ticketing/oauth/grant_type_tokens/
func ExchangeClientCredentials(ctx context.Context, httpClient *http.Client, baseURL, clientID, clientSecret, scope string) (string, error) {
	payload, err := json.Marshal(clientCredentialsRequest{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/oauth/tokens", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchanging oauth client credentials: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("exchanging oauth client credentials: %s: %s", resp.Status, body)
	}

	var result clientCredentialsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("exchanging oauth client credentials: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("exchanging oauth client credentials: response carried no access_token")
	}

	return result.AccessToken, nil
}
