package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/nukosuke/go-zendesk/zendesk"
)

func newTestClient(server *httptest.Server, credential zendesk.Credential) *Client {
	return NewClient(zendesk.Client{}, server.URL+"/api/v2", credential, server.Client())
}

func TestPatchWebhookSendsOnlyTheGivenFields(t *testing.T) {
	var method, path, contentType string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path, contentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := newTestClient(server, zendesk.NewBearerTokenCredential("token")).
		PatchWebhook(context.Background(), "01ABC", map[string]any{"description": "new"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != http.MethodPatch || path != "/api/v2/webhooks/01ABC" || contentType != "application/json" {
		t.Fatalf("sent %s %s as %q, should have sent PATCH /api/v2/webhooks/01ABC as application/json", method, path, contentType)
	}
	want := map[string]any{"webhook": map[string]any{"description": "new"}}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body was %v, should have been %v", body, want)
	}
}

func TestPatchWebhookAuthenticatesWithABearerToken(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	newTestClient(server, zendesk.NewBearerTokenCredential("issued-token")).
		PatchWebhook(context.Background(), "01ABC", map[string]any{"description": "new"})

	if authorization != "Bearer issued-token" {
		t.Fatalf("Authorization was %q, should have been %q", authorization, "Bearer issued-token")
	}
}

func TestPatchWebhookAuthenticatesWithAnAPIToken(t *testing.T) {
	var user, password string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, _ = r.BasicAuth()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	credential := zendesk.NewAPITokenCredential("agent@example.com", "api-token")
	newTestClient(server, credential).PatchWebhook(context.Background(), "01ABC", map[string]any{"description": "new"})

	if user != credential.Email() || password != "api-token" {
		t.Fatalf("basic auth was %q:%q, should have been %q:%q", user, password, credential.Email(), "api-token")
	}
}

func TestPatchWebhookReturnsTheResponseOfARejectedPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"code":"InvalidWebhookAuthentication"}]}`))
	}))
	defer server.Close()

	err := newTestClient(server, zendesk.NewBearerTokenCredential("token")).
		PatchWebhook(context.Background(), "01ABC", map[string]any{"description": "new"})

	if err == nil || !strings.Contains(err.Error(), "400") || !strings.Contains(err.Error(), "InvalidWebhookAuthentication") {
		t.Fatalf("error was %v, should have carried the status and the response body", err)
	}
}
