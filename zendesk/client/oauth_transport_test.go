package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestOAuthTransportDoesNotExchangeBeforeTheFirstRequest(t *testing.T) {
	var exchanges int32
	exchange := func(ctx context.Context) (string, error) {
		atomic.AddInt32(&exchanges, 1)
		return "issued-token", nil
	}

	NewOAuthTransport(nil, exchange)

	if exchanges != 0 {
		t.Fatalf("exchanged %d times, should not have exchanged", exchanges)
	}
}

func TestOAuthTransportSendsTheExchangedTokenAsBearer(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
	}))
	defer server.Close()
	exchange := func(ctx context.Context) (string, error) { return "issued-token", nil }
	httpClient := &http.Client{Transport: NewOAuthTransport(nil, exchange)}
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	req.Header.Set("Authorization", "Bearer ")

	resp, err := httpClient.Do(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if authorization != "Bearer issued-token" {
		t.Fatalf("Authorization was %q, should have been %q", authorization, "Bearer issued-token")
	}
}

func TestOAuthTransportExchangesOnceAcrossRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	var exchanges int32
	exchange := func(ctx context.Context) (string, error) {
		atomic.AddInt32(&exchanges, 1)
		return "issued-token", nil
	}
	httpClient := &http.Client{Transport: NewOAuthTransport(nil, exchange)}

	for range 2 {
		resp, err := httpClient.Get(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp.Body.Close()
	}

	if exchanges != 1 {
		t.Fatalf("exchanged %d times, should have exchanged once", exchanges)
	}
}

func TestOAuthTransportFailsTheRequestWhenTheExchangeFails(t *testing.T) {
	var reached bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer server.Close()
	exchange := func(ctx context.Context) (string, error) {
		return "", errors.New("exchanging oauth client credentials: 401 Unauthorized")
	}
	httpClient := &http.Client{Transport: NewOAuthTransport(nil, exchange)}

	_, err := httpClient.Get(server.URL)

	if err == nil || !strings.Contains(err.Error(), "exchanging oauth client credentials") {
		t.Fatalf("error was %v, should have named the failed exchange", err)
	}
	if reached {
		t.Fatal("the request reached the server, should have stopped at the failed exchange")
	}
}
