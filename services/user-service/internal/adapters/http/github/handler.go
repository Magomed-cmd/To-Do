package github

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	authadapter "todoapp/services/user-service/internal/adapters/auth"
	"todoapp/services/user-service/internal/adapters/http/common"
	"todoapp/services/user-service/internal/dto"
	"todoapp/services/user-service/internal/ports"
)

type oauthProvider struct {
	generateState func() (string, error)
	authCodeURL   func(state string) string
	exchange      func(ctx context.Context, code string) (*oauth2.Token, error)
	fetchUser     func(ctx context.Context, token *oauth2.Token) (*authadapter.GitHubProfile, error)
}

type Handler struct {
	oauth           oauthProvider
	service         ports.UserService
	stateCookieName string
	cookieTTL       time.Duration
}

func New(oauth *authadapter.GitHubOAuth, service ports.UserService) *Handler {
	if oauth == nil {
		panic("github oauth provider is nil")
	}
	return newHandler(oauthProvider{
		generateState: oauth.GenerateState,
		authCodeURL:   oauth.AuthCodeURL,
		exchange:      oauth.Exchange,
		fetchUser:     oauth.FetchUser,
	}, service)
}

func newHandler(oauth oauthProvider, service ports.UserService) *Handler {
	return &Handler{
		oauth:           oauth,
		service:         service,
		stateCookieName: "github_oauth_state",
		cookieTTL:       5 * time.Minute,
	}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("/auth/github", h.Begin)
	router.GET("/auth/github/callback", h.Callback)
}

func (h *Handler) Begin(ctx *gin.Context) {
	state, err := h.oauth.generateState()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "STATE_GENERATION_FAILED"})
		return
	}
	ctx.SetCookie(h.stateCookieName, state, int(h.cookieTTL.Seconds()), "/", "", false, true)
	ctx.Redirect(http.StatusFound, h.oauth.authCodeURL(state))
}

func (h *Handler) Callback(ctx *gin.Context) {
	state := ctx.Query("state")
	code := ctx.Query("code")
	if state == "" || code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "MISSING_PARAMETERS"})
		return
	}
	stored, err := ctx.Cookie(h.stateCookieName)
	if err != nil || stored != state {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_STATE"})
		return
	}
	token, err := h.oauth.exchange(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "OAUTH_EXCHANGE_FAILED"})
		return
	}
	profile, err := h.oauth.fetchUser(ctx.Request.Context(), token)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "OAUTH_USER_FETCH_FAILED"})
		return
	}
	result, err := h.service.GitHubLogin(ctx.Request.Context(), ports.OAuthLoginInput{
		ProviderID: profile.ID,
		Email:      profile.Email,
		Name:       profile.Name,
		AvatarURL:  profile.AvatarURL,
	})
	if err != nil {
		common.WriteDomainError(ctx, err)
		return
	}
	ctx.SetCookie(h.stateCookieName, "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, dto.NewAuthResponse(*result))
}
