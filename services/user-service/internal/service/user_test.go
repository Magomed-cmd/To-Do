package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"todoapp/services/user-service/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"

	"todoapp/services/user-service/internal/domain"
	"todoapp/services/user-service/internal/ports"
)

type repoStub struct {
	userByEmail  *entities.User
	userByID     *entities.User
	createErr    error
	getErr       error
	updateErr    error
	deleteErr    error
	created      bool
	updated      bool
	prefs        *entities.UserPreferences
	prefErr      error
	upsertErr    error
	upserted     bool
	listFunc     func(ctx context.Context, limit, offset int) ([]entities.User, error)
	withTxCalled bool
}

func (r *repoStub) Create(ctx context.Context, user *entities.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	user.ID = 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.created = true
	return nil
}

func (r *repoStub) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.userByID == nil {
		return nil, domain.ErrUserNotFound
	}
	return r.userByID, nil
}

func (r *repoStub) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.userByEmail == nil {
		return nil, domain.ErrUserNotFound
	}
	return r.userByEmail, nil
}

func (r *repoStub) Update(ctx context.Context, user *entities.User) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updated = true
	return nil
}

func (r *repoStub) Delete(ctx context.Context, id int64) error {
	return r.deleteErr
}

func (r *repoStub) UpsertPreferences(ctx context.Context, prefs entities.UserPreferences) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upserted = true
	r.prefs = &prefs
	return nil
}

func (r *repoStub) GetPreferences(ctx context.Context, userID int64) (*entities.UserPreferences, error) {
	if r.prefErr != nil {
		return nil, r.prefErr
	}
	if r.prefs == nil {
		return &entities.UserPreferences{UserID: userID}, nil
	}
	return r.prefs, nil
}

func (r *repoStub) List(ctx context.Context, limit, offset int) ([]entities.User, error) {
	if r.listFunc != nil {
		return r.listFunc(ctx, limit, offset)
	}
	return []entities.User{}, nil
}

func (r *repoStub) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	r.withTxCalled = true
	return fn(ctx)
}

type tokenManagerStub struct {
	expectPayload ports.TokenPayload
	parseClaims   *ports.TokenClaims
	parseErr      error
}

func (t *tokenManagerStub) GenerateAccessToken(payload ports.TokenPayload) (string, time.Time, error) {
	t.expectPayload = payload
	return "access", time.Now().Add(time.Minute), nil
}

func (t *tokenManagerStub) GenerateRefreshToken(payload ports.TokenPayload, tokenID string) (string, time.Time, error) {
	t.expectPayload = payload
	return "refresh", time.Now().Add(time.Hour), nil
}

func (t *tokenManagerStub) ParseAccessToken(token string) (*ports.TokenClaims, error) {
	return t.parseClaims, t.parseErr
}

func (t *tokenManagerStub) ParseRefreshToken(token string) (*ports.TokenClaims, error) {
	if t.parseErr != nil {
		return nil, t.parseErr
	}
	return t.parseClaims, nil
}

func TestRegister(t *testing.T) {
	repo := &repoStub{}
	tokens := &tokenManagerStub{}
	svc := NewUserService(repo, tokens)

	result, err := svc.Register(context.Background(), ports.RegisterInput{
		Email:    "new@example.com",
		Name:     "New User",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !repo.created {
		t.Fatalf("expected create invocation")
	}
	if result.User.ID == 0 {
		t.Fatalf("expected user id")
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatalf("expected tokens")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	repo := &repoStub{
		userByEmail: &entities.User{ID: 1},
	}
	svc := NewUserService(repo, &tokenManagerStub{})
	_, err := svc.Register(context.Background(), ports.RegisterInput{
		Email:    "existing@example.com",
		Name:     "Existing",
		Password: "password",
	})
	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("expected duplicate error")
	}
}

func TestLoginSuccess(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	user := &entities.User{
		ID:           1,
		Email:        "user@example.com",
		IsActive:     true,
		PasswordHash: string(hash),
	}
	repo := &repoStub{userByEmail: user}
	tokens := &tokenManagerStub{}
	svc := NewUserService(repo, tokens)
	result, err := svc.Login(context.Background(), ports.LoginInput{
		Email:    "user@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if result.User.ID != 1 {
		t.Fatalf("expected user id 1")
	}
	if tokens.expectPayload.UserID != 1 {
		t.Fatalf("expected payload recorded")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	user := &entities.User{
		ID:           1,
		Email:        "user@example.com",
		IsActive:     true,
		PasswordHash: string(hash),
	}
	repo := &repoStub{userByEmail: user}
	svc := NewUserService(repo, &tokenManagerStub{})
	_, err := svc.Login(context.Background(), ports.LoginInput{
		Email:    "user@example.com",
		Password: "wrong",
	})
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials error")
	}
}

func TestRefreshToken(t *testing.T) {
	user := &entities.User{
		ID:       1,
		Email:    "user@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo := &repoStub{userByID: user}
	tokens := &tokenManagerStub{
		parseClaims: &ports.TokenClaims{
			UserID:  1,
			Email:   "user@example.com",
			Role:    "user",
			TokenID: "abc",
		},
	}
	svc := NewUserService(repo, tokens)
	result, err := svc.RefreshToken(context.Background(), "refresh")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if result.RefreshToken == "" || result.AccessToken == "" {
		t.Fatalf("expected new tokens")
	}
}

func TestRefreshTokenParseError(t *testing.T) {
	repo := &repoStub{}
	tokens := &tokenManagerStub{
		parseErr: errors.New("boom"),
	}
	svc := NewUserService(repo, tokens)
	_, err := svc.RefreshToken(context.Background(), "refresh")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGitHubLoginCreatesUser(t *testing.T) {
	repo := &repoStub{}
	tokens := &tokenManagerStub{}
	svc := NewUserService(repo, tokens)
	result, err := svc.GitHubLogin(context.Background(), ports.OAuthLoginInput{
		ProviderID: 100,
		Email:      "oauth@example.com",
		Name:       "OAuth User",
		AvatarURL:  "https://avatar",
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !repo.created {
		t.Fatalf("expected create call")
	}
	if result.User.GitHubID == nil || *result.User.GitHubID != 100 {
		t.Fatalf("expected github id to be set")
	}
}

func TestGitHubLoginUpdatesExistingUser(t *testing.T) {
	gitID := int64(50)
	repo := &repoStub{
		userByEmail: &entities.User{ID: 1, Email: "oauth@example.com", IsActive: true},
	}
	tokens := &tokenManagerStub{}
	svc := NewUserService(repo, tokens)
	_, err := svc.GitHubLogin(context.Background(), ports.OAuthLoginInput{
		ProviderID: gitID,
		Email:      "oauth@example.com",
		Name:       "OAuth User",
		AvatarURL:  "https://avatar",
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !repo.updated {
		t.Fatalf("expected update call")
	}
	if repo.userByEmail.GitHubID == nil || *repo.userByEmail.GitHubID != gitID {
		t.Fatalf("expected github id set on user")
	}
}

func TestGetProfile(t *testing.T) {
	repo := &repoStub{userByID: &entities.User{ID: 1}}
	svc := NewUserService(repo, &tokenManagerStub{})
	user, err := svc.GetProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if user.ID != 1 {
		t.Fatalf("expected user id 1")
	}
}

func TestUpdateProfile(t *testing.T) {
	name := "Updated"
	avatar := "https://avatar"
	repo := &repoStub{userByID: &entities.User{ID: 1, Name: "Old"}}
	svc := NewUserService(repo, &tokenManagerStub{})
	user, err := svc.UpdateProfile(context.Background(), 1, ports.UpdateProfileInput{Name: &name, AvatarURL: &avatar})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if user.Name != "Updated" || user.AvatarURL != avatar {
		t.Fatalf("profile not updated")
	}
	if !repo.updated {
		t.Fatalf("expected update to persist")
	}
}

func TestGetPreferences(t *testing.T) {
	prefs := &entities.UserPreferences{UserID: 1, Theme: "dark"}
	repo := &repoStub{prefs: prefs}
	svc := NewUserService(repo, &tokenManagerStub{})
	out, err := svc.GetPreferences(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if out.Theme != "dark" {
		t.Fatalf("unexpected theme")
	}
}

func TestUpdatePreferences(t *testing.T) {
	repo := &repoStub{prefs: &entities.UserPreferences{UserID: 1}}
	svc := NewUserService(repo, &tokenManagerStub{})
	newTheme := "dark"
	enabled := true
	prefs, err := svc.UpdatePreferences(context.Background(), 1, ports.UpdatePreferencesInput{Theme: &newTheme, NotificationsEnabled: &enabled})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if prefs.Theme != "dark" || !prefs.NotificationsEnabled {
		t.Fatalf("preferences not updated")
	}
	if !repo.upserted {
		t.Fatalf("expected upsert call")
	}
}

func TestListUsers(t *testing.T) {
	repo := &repoStub{userByID: &entities.User{ID: 1}}
	repoList := func(ctx context.Context, limit, offset int) ([]entities.User, error) {
		return []entities.User{{ID: 1}}, nil
	}
	repo.listFunc = repoList
	svc := NewUserService(repo, &tokenManagerStub{})
	users, err := svc.ListUsers(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected one user")
	}
}

func TestUpdateUserRole(t *testing.T) {
	repo := &repoStub{userByID: &entities.User{ID: 1, Role: "user"}}
	svc := NewUserService(repo, &tokenManagerStub{})
	user, err := svc.UpdateUserRole(context.Background(), 1, "admin")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if user.Role != "admin" {
		t.Fatalf("role not updated")
	}
}

func TestUpdateUserStatus(t *testing.T) {
	repo := &repoStub{userByID: &entities.User{ID: 1, IsActive: false}}
	svc := NewUserService(repo, &tokenManagerStub{})
	user, err := svc.UpdateUserStatus(context.Background(), 1, true)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !user.IsActive {
		t.Fatalf("status not updated")
	}
}
