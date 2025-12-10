package auth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

type roundTripper func(req *http.Request) (*http.Response, error)

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

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
	transport := roundTripper(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/token":
			_ = req.ParseForm()
			if req.FormValue("code") != "code" {
				t.Fatalf("unexpected code %s", req.FormValue("code"))
			}
			body := io.NopCloser(strings.NewReader(`{"access_token":"token","token_type":"bearer","expires_in":3600}`))
			return &http.Response{StatusCode: http.StatusOK, Body: body, Header: http.Header{"Content-Type": []string{"application/json"}}}, nil
		case "/user":
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("missing token header")
			}
			body := io.NopCloser(strings.NewReader(`{"id":123,"email":"user@example.com","name":"User","avatar_url":"https://avatar"}`))
			return &http.Response{StatusCode: http.StatusOK, Body: body, Header: http.Header{"Content-Type": []string{"application/json"}}}, nil
		default:
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
		}
	})

	oauth := NewGitHubOAuth("id", "secret", "http://localhost/callback", []string{"user"})
	oauth.config.Endpoint = oauth2.Endpoint{AuthURL: "https://example.com/authorize", TokenURL: "https://example.com/token"}
	oauth.WithHTTPClient(&http.Client{Transport: transport})
	oauth.WithAPIBase("https://example.com")
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
