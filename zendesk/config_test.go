package zendesk

import "testing"

func TestConfigUsesOAuth(t *testing.T) {
	cases := []struct {
		name      string
		config    Config
		wantOAuth bool
		wantErr   bool
	}{
		{"api token", Config{Email: "agent@example.com", Token: "token"}, false, false},
		{"oauth client", Config{OAuthClientID: "id", OAuthClientSecret: "secret"}, true, false},
		{"oauth client with scope", Config{OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScope: "read"}, true, false},
		{"both kinds", Config{Email: "agent@example.com", Token: "token", OAuthClientID: "id", OAuthClientSecret: "secret"}, false, true},
		{"neither kind", Config{}, false, true},
		{"scope alone", Config{OAuthScope: "read"}, false, true},
		{"email without token", Config{Email: "agent@example.com"}, false, true},
		{"token without email", Config{Token: "token"}, false, true},
		{"client id without secret", Config{OAuthClientID: "id"}, false, true},
		{"secret without client id", Config{OAuthClientSecret: "secret"}, false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oauth, err := c.config.usesOAuth()

			if (err != nil) != c.wantErr {
				t.Fatalf("error was %v, wantErr %v", err, c.wantErr)
			}
			if oauth != c.wantOAuth {
				t.Fatalf("usesOAuth was %v, should have been %v", oauth, c.wantOAuth)
			}
		})
	}
}
