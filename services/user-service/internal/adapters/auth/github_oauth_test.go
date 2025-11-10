package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestGenerateState(t *testing.T) {
	oauth := NewGitHubOAuth("id", "secret", "http://localhost/callback", []string{"user"})
	state, err := oauth.GenerateState()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(state) == 0 {
		t.Fatalf("expected non-empty state")
	}
}

func TestAuthCodeURLIncludesState(t *testing.T) {
	oauth := NewGitHubOAuth("id", "secret", "http://localhost/callback", []string{"user"})
	url := oauth.AuthCodeURL("abc")
	if !strings.Contains(url, "state=abc") {
		t.Fatalf("expected state parameter in url")
	}
}

func TestExchangeAndFetchUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_ = r.ParseForm()
			if r.FormValue("code") != "code" {
				t.Fatalf("unexpected code %s", r.FormValue("code"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"token","token_type":"bearer","expires_in":3600}`))
		case "/user":
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("missing token header")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":123,"email":"user@example.com","name":"User","avatar_url":"https://avatar"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	oauth := NewGitHubOAuth("id", "secret", "http://localhost/callback", []string{"user"})
	oauth.config.Endpoint = oauth2.Endpoint{AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token"}
	oauth.WithHTTPClient(server.Client())
	oauth.WithAPIBase(server.URL)
	oauth.WithStateGenerator(func() (string, error) { return "state", nil })

	token, err := oauth.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if token.AccessToken != "token" {
		t.Fatalf("unexpected access token")
	}

	profile, err := oauth.FetchUser(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if profile.ID != 123 || profile.Email != "user@example.com" {
		t.Fatalf("unexpected profile")
	}
}
