package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubOAuth struct {
	config         *oauth2.Config
	httpClient     *http.Client
	apiBase        string
	stateGenerator func() (string, error)
}

type GitHubProfile struct {
	ID        int64
	Email     string
	Name      string
	AvatarURL string
}

func NewGitHubOAuth(clientID, clientSecret, redirectURL string, scopes []string) *GitHubOAuth {
	return &GitHubOAuth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		},
		httpClient:     http.DefaultClient,
		apiBase:        "https://api.github.com",
		stateGenerator: generateState,
	}
}

func (g *GitHubOAuth) WithHTTPClient(client *http.Client) {
	if client != nil {
		g.httpClient = client
	}
}

func (g *GitHubOAuth) WithAPIBase(apiBase string) {
	if apiBase != "" {
		g.apiBase = apiBase
	}
}

func (g *GitHubOAuth) WithStateGenerator(generator func() (string, error)) {
	if generator != nil {
		g.stateGenerator = generator
	}
}

func (g *GitHubOAuth) GenerateState() (string, error) {
	return g.stateGenerator()
}

func (g *GitHubOAuth) AuthCodeURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GitHubOAuth) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	ctx = g.contextWithClient(ctx)
	return g.config.Exchange(ctx, code)
}

func (g *GitHubOAuth) FetchUser(ctx context.Context, token *oauth2.Token) (*GitHubProfile, error) {
	client := g.oauthClient(ctx, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/user", g.apiBase), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github user request failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	profile := GitHubProfile{
		ID:        payload.ID,
		Email:     payload.Email,
		Name:      payload.Name,
		AvatarURL: payload.AvatarURL,
	}
	return &profile, nil
}

func (g *GitHubOAuth) oauthClient(ctx context.Context, token *oauth2.Token) *http.Client {
	ctx = g.contextWithClient(ctx)
	return g.config.Client(ctx, token)
}

func (g *GitHubOAuth) contextWithClient(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, oauth2.HTTPClient, g.httpClient)
}

func generateState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
