package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

type Service struct {
	repository *Repository
	jwtSecret  []byte
	tokenTTL   time.Duration
}

func NewService(repository *Repository, jwtSecret string) *Service {
	return &Service{
		repository: repository,
		jwtSecret:  []byte(jwtSecret),
		tokenTTL:   24 * time.Hour,
	}
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (LoginResponse, error) {
	user, err := s.repository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}

	if !user.IsActive {
		return LoginResponse{}, ErrInactiveUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	token, expiresAt, err := s.issueToken(user)
	if err != nil {
		return LoginResponse{}, err
	}

	_ = s.repository.TouchLastLogin(ctx, user.ID)

	return LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		User:        user,
	}, nil
}

func (s *Service) Me(ctx context.Context, userID string) (User, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	if !user.IsActive {
		return User{}, ErrInactiveUser
	}
	return user, nil
}

func (s *Service) ParseToken(tokenString string) (Claims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return Claims{}, ErrInvalidToken
	}

	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)
	organizationID, _ := claims["organizationId"].(string)

	return Claims{
		UserID:         userID,
		Email:          email,
		Role:           role,
		OrganizationID: organizationID,
	}, nil
}

func (s *Service) issueToken(user User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenTTL)
	organizationID := ""
	if user.OrganizationID != nil {
		organizationID = *user.OrganizationID
	}

	claims := jwt.MapClaims{
		"sub":            user.ID,
		"email":          user.Email,
		"role":           user.PrimaryRole,
		"organizationId": organizationID,
		"iat":            time.Now().Unix(),
		"exp":            expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
