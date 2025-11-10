package service

import (
	"context"
	"errors"
	"time"

	"todoapp/services/user-service/internal/domain"
	"todoapp/services/user-service/internal/domain/entities"
	"todoapp/services/user-service/internal/ports"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultUserRole = "user"
)

type UserService struct {
	repo   ports.UserRepository
	tokens ports.TokenManager
	now    func() time.Time
}

func NewUserService(repo ports.UserRepository, tokens ports.TokenManager) *UserService {
	return &UserService{
		repo:   repo,
		tokens: tokens,
		now:    time.Now,
	}
}

var _ ports.UserService = (*UserService)(nil)

func (s *UserService) Register(ctx context.Context, input ports.RegisterInput) (*ports.AuthResult, error) {
	_, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	user, err := s.createUserWithPassword(input.Email, input.Name, input.Password)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.buildAuthResult(user)
}

func (s *UserService) Login(ctx context.Context, input ports.LoginInput) (*ports.AuthResult, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if err := s.validateUserActive(user); err != nil {
		return nil, err
	}

	if err := s.validatePassword(user.PasswordHash, input.Password); err != nil {
		return nil, err
	}

	return s.buildAuthResult(user)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*ports.AuthTokens, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.validateUserActive(user); err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

func (s *UserService) GitHubLogin(ctx context.Context, input ports.OAuthLoginInput) (*ports.AuthResult, error) {
	user, err := s.findOrCreateGitHubUser(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.validateUserActive(user); err != nil {
		return nil, err
	}

	return s.buildAuthResult(user)
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*entities.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, input ports.UpdateProfileInput) (*entities.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.applyProfileUpdates(user, input)

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetPreferences(ctx context.Context, userID int64) (*entities.UserPreferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *UserService) UpdatePreferences(ctx context.Context, userID int64, input ports.UpdatePreferencesInput) (*entities.UserPreferences, error) {
	current, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.applyPreferencesUpdates(current, input)
	current.UserID = userID

	if err := s.repo.UpsertPreferences(ctx, *current); err != nil {
		return nil, err
	}

	return current, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]entities.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *UserService) UpdateUserRole(ctx context.Context, userID int64, role string) (*entities.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Role = role

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, userID int64, isActive bool) (*entities.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.IsActive = isActive

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) WithNow(now func() time.Time) {
	s.now = now
}

func (s *UserService) createUserWithPassword(email, name, password string) (*entities.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &entities.User{
		Email:        email,
		Name:         name,
		Role:         DefaultUserRole,
		IsActive:     true,
		PasswordHash: string(hash),
	}, nil
}

func (s *UserService) validateUserActive(user *entities.User) error {
	if !user.IsActive {
		return domain.ErrUserInactive
	}
	return nil
}

func (s *UserService) validatePassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return domain.ErrInvalidCredentials
	}
	return nil
}

func (s *UserService) buildAuthResult(user *entities.User) (*ports.AuthResult, error) {
	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &ports.AuthResult{
		User:   *user,
		Tokens: *tokens,
	}, nil
}

func (s *UserService) generateTokens(user *entities.User) (*ports.AuthTokens, error) {
	payload := ports.TokenPayload{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	access, accessExp, err := s.tokens.GenerateAccessToken(payload)
	if err != nil {
		return nil, err
	}

	refreshID := uuid.NewString()
	refresh, refreshExp, err := s.tokens.GenerateRefreshToken(payload, refreshID)
	if err != nil {
		return nil, err
	}

	return &ports.AuthTokens{
		AccessToken:           access,
		AccessTokenExpiresAt:  accessExp,
		RefreshToken:          refresh,
		RefreshTokenExpiresAt: refreshExp,
	}, nil
}

func (s *UserService) findOrCreateGitHubUser(ctx context.Context, input ports.OAuthLoginInput) (*entities.User, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)

	if err != nil && errors.Is(err, domain.ErrUserNotFound) {
		return s.createGitHubUser(ctx, input)
	}

	if err != nil {
		return nil, err
	}

	if s.shouldUpdateGitHubUser(user, input) {
		s.updateGitHubUserInfo(user, input)
		if err := s.repo.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) createGitHubUser(ctx context.Context, input ports.OAuthLoginInput) (*entities.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Email:        input.Email,
		Name:         input.Name,
		AvatarURL:    input.AvatarURL,
		Role:         DefaultUserRole,
		IsActive:     true,
		PasswordHash: string(hash),
		GitHubID:     &input.ProviderID,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) shouldUpdateGitHubUser(user *entities.User, input ports.OAuthLoginInput) bool {
	return user.GitHubID == nil ||
		*user.GitHubID != input.ProviderID ||
		user.AvatarURL != input.AvatarURL ||
		user.Name != input.Name
}

func (s *UserService) updateGitHubUserInfo(user *entities.User, input ports.OAuthLoginInput) {
	user.GitHubID = &input.ProviderID
	user.AvatarURL = input.AvatarURL
	user.Name = input.Name
}

func (s *UserService) applyProfileUpdates(user *entities.User, input ports.UpdateProfileInput) {
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}
}

func (s *UserService) applyPreferencesUpdates(prefs *entities.UserPreferences, input ports.UpdatePreferencesInput) {
	if input.NotificationsEnabled != nil {
		prefs.NotificationsEnabled = *input.NotificationsEnabled
	}
	if input.EmailNotifications != nil {
		prefs.EmailNotifications = *input.EmailNotifications
	}
	if input.Theme != nil {
		prefs.Theme = *input.Theme
	}
	if input.Language != nil {
		prefs.Language = *input.Language
	}
	if input.Timezone != nil {
		prefs.Timezone = *input.Timezone
	}
}
