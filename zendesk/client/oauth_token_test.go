package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExchangeClientCredentialsReturnsAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"access_token":"issued-token","token_type":"bearer","expires_in":1799}`))
	}))
	defer server.Close()

	token, err := ExchangeClientCredentials(context.Background(), server.Client(), server.URL, "id", "secret", "read")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "issued-token" {
		t.Fatalf("token was %q, should have been %q", token, "issued-token")
	}
}

func TestExchangeClientCredentialsPostsTheGrantToTheTokenEndpoint(t *testing.T) {
	var method, path, contentType string
	var body map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path, contentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		json.NewDecoder(r.Body).Decode(&body)
		w.Write([]byte(`{"access_token":"issued-token"}`))
	}))
	defer server.Close()

	ExchangeClientCredentials(context.Background(), server.Client(), server.URL, "the-id", "the-secret", "read write")

	if method != http.MethodPost || path != "/oauth/tokens" || contentType != "application/json" {
		t.Fatalf("request was %s %s (%s), should have been POST /oauth/tokens (application/json)", method, path, contentType)
	}
	expected := map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     "the-id",
		"client_secret": "the-secret",
		"scope":         "read write",
	}
	for key, value := range expected {
		if body[key] != value {
			t.Fatalf("body %s was %v, should have been %v", key, body[key], value)
		}
	}
}

func TestExchangeClientCredentialsOmitsAnEmptyScope(t *testing.T) {
	var body map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Write([]byte(`{"access_token":"issued-token"}`))
	}))
	defer server.Close()

	ExchangeClientCredentials(context.Background(), server.Client(), server.URL, "id", "secret", "")

	if _, present := body["scope"]; present {
		t.Fatalf("body carried scope %v, should have left it out", body["scope"])
	}
}

func TestExchangeClientCredentialsReportsARejectionWithoutTheSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Couldn't authenticate you"}`))
	}))
	defer server.Close()

	_, err := ExchangeClientCredentials(context.Background(), server.Client(), server.URL, "id", "do-not-leak", "")

	if err == nil {
		t.Fatalf("expected an error for a 401")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Couldn't authenticate you") {
		t.Fatalf("error %q should carry the status and Zendesk's message", err)
	}
	if strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("error %q leaks the client secret", err)
	}
}

func TestExchangeClientCredentialsRejectsAResponseWithoutAToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"token_type":"bearer"}`))
	}))
	defer server.Close()

	_, err := ExchangeClientCredentials(context.Background(), server.Client(), server.URL, "id", "secret", "")

	if err == nil {
		t.Fatalf("expected an error for a response without access_token")
	}
}
