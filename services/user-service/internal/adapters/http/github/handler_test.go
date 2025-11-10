package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	authadapter "todoapp/services/user-service/internal/adapters/auth"
	authmocks "todoapp/services/user-service/internal/adapters/http/auth"
	"todoapp/services/user-service/internal/domain/entities"
	"todoapp/services/user-service/internal/ports"
)

func TestBegin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oauth := newStubOAuthProvider(t, nil)
	service := authmocks.NewMockUserService(t)
	handler := newHandler(oauth, service)
	router := gin.New()
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodGet, "/auth/github", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusFound, res.Code)
}

func TestCallbackSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := &oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}
	profile := &authadapter.GitHubProfile{ID: 1, Email: "user@example.com", Name: "User", AvatarURL: "https://avatar"}
	oauth := newStubOAuthProvider(t, func(p *oauthProvider) {
		p.exchange = func(ctx context.Context, code string) (*oauth2.Token, error) {
			assert.NotNil(t, ctx)
			assert.Equal(t, "code", code)
			return token, nil
		}
		p.fetchUser = func(ctx context.Context, tok *oauth2.Token) (*authadapter.GitHubProfile, error) {
			assert.NotNil(t, ctx)
			assert.Equal(t, token, tok)
			return profile, nil
		}
	})
	service := authmocks.NewMockUserService(t)
	service.On("GitHubLogin", mock.Anything, ports.OAuthLoginInput{
		ProviderID: profile.ID,
		Email:      profile.Email,
		Name:       profile.Name,
		AvatarURL:  profile.AvatarURL,
	}).Return(&ports.AuthResult{User: entities.User{ID: 1}}, nil)
	handler := newHandler(oauth, service)
	router := gin.New()
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=code&state=state", nil)
	req.AddCookie(&http.Cookie{Name: "github_oauth_state", Value: "state"})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	service.AssertExpectations(t)
}

func TestCallbackInvalidState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oauth := newStubOAuthProvider(t, nil)
	service := authmocks.NewMockUserService(t)
	handler := newHandler(oauth, service)
	router := gin.New()
	handler.RegisterRoutes(router)
	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=code&state=other", nil)
	req.AddCookie(&http.Cookie{Name: "github_oauth_state", Value: "state"})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func newStubOAuthProvider(t *testing.T, customize func(*oauthProvider)) oauthProvider {
	t.Helper()

	provider := oauthProvider{
		generateState: func() (string, error) {
			return "state", nil
		},
		authCodeURL: func(state string) string {
			assert.Equal(t, "state", state)
			return "https://github.com/login"
		},
		exchange: func(ctx context.Context, code string) (*oauth2.Token, error) {
			t.Fatalf("unexpected Exchange call with code %s", code)
			return nil, nil
		},
		fetchUser: func(ctx context.Context, token *oauth2.Token) (*authadapter.GitHubProfile, error) {
			t.Fatalf("unexpected FetchUser call with token %v", token)
			return nil, nil
		},
	}

	if customize != nil {
		customize(&provider)
	}

	return provider
}
