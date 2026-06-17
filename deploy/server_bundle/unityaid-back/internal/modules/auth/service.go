package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrEmailNotVerified   = errors.New("email is not verified")
	ErrInvalidToken       = errors.New("invalid token")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrTooManyAttempts    = errors.New("too many login attempts")
)

type Service struct {
	repository      *Repository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	exposeDevTokens bool
	limiter         *loginLimiter
}

func NewService(repository *Repository, jwtSecret string, exposeDevTokens bool) *Service {
	return &Service{
		repository:      repository,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 30 * 24 * time.Hour,
		exposeDevTokens: exposeDevTokens,
		limiter:         newLoginLimiter(5, 15*time.Minute),
	}
}

func (s *Service) Register(ctx context.Context, request RegisterRequest) (DevTokenResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return DevTokenResponse{}, err
	}

	user, err := s.repository.CreateUser(ctx, request, string(passwordHash))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return DevTokenResponse{}, ErrEmailAlreadyExists
		}
		return DevTokenResponse{}, err
	}

	token, err := randomToken()
	if err != nil {
		return DevTokenResponse{}, err
	}
	if err := s.repository.CreateEmailVerificationToken(ctx, user.ID, tokenHash(token), time.Now().Add(24*time.Hour)); err != nil {
		return DevTokenResponse{}, err
	}

	response := DevTokenResponse{Status: "verification_required"}
	if s.exposeDevTokens {
		response.Token = token
	}
	return response, nil
}

func (s *Service) Login(ctx context.Context, request LoginRequest, userAgent string, ipAddress string) (LoginResponse, error) {
	limitKey := strings.ToLower(strings.TrimSpace(request.Email)) + "|" + ipAddress
	if !s.limiter.Allow(limitKey) {
		return LoginResponse{}, ErrTooManyAttempts
	}

	user, err := s.repository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.limiter.Fail(limitKey)
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}

	if !user.IsActive {
		return LoginResponse{}, ErrInactiveUser
	}
	if !user.IsEmailVerified {
		return LoginResponse{}, ErrEmailNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		s.limiter.Fail(limitKey)
		return LoginResponse{}, ErrInvalidCredentials
	}

	response, err := s.issueAuthTokens(ctx, user, userAgent, ipAddress)
	if err != nil {
		return LoginResponse{}, err
	}

	s.limiter.Reset(limitKey)
	_ = s.repository.TouchLastLogin(ctx, user.ID)

	return response, nil
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

func (s *Service) Refresh(ctx context.Context, request RefreshRequest, userAgent string, ipAddress string) (LoginResponse, error) {
	hash := tokenHash(request.RefreshToken)
	user, err := s.repository.FindUserByRefreshTokenHash(ctx, hash)
	if err != nil {
		return LoginResponse{}, ErrInvalidToken
	}
	if !user.IsActive {
		return LoginResponse{}, ErrInactiveUser
	}
	if !user.IsEmailVerified {
		return LoginResponse{}, ErrEmailNotVerified
	}

	if err := s.repository.RevokeRefreshSession(ctx, hash); err != nil {
		return LoginResponse{}, err
	}

	return s.issueAuthTokens(ctx, user, userAgent, ipAddress)
}

func (s *Service) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	if accessToken != "" {
		claims, err := s.ParseToken(ctx, accessToken)
		if err == nil {
			if err := s.repository.RevokeAccessToken(ctx, claims.JTI, claims.UserID, claims.ExpiresAt); err != nil {
				return err
			}
		}
	}

	if refreshToken != "" {
		return s.repository.RevokeRefreshSession(ctx, tokenHash(refreshToken))
	}

	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, request ForgotPasswordRequest) (DevTokenResponse, error) {
	user, err := s.repository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return DevTokenResponse{Status: "ok"}, nil
		}
		return DevTokenResponse{}, err
	}

	token, err := randomToken()
	if err != nil {
		return DevTokenResponse{}, err
	}
	if err := s.repository.CreatePasswordResetToken(ctx, user.ID, tokenHash(token), time.Now().Add(time.Hour)); err != nil {
		return DevTokenResponse{}, err
	}
	response := DevTokenResponse{Status: "ok"}
	if s.exposeDevTokens {
		response.Token = token
	}
	return response, nil
}

func (s *Service) ResetPassword(ctx context.Context, request ResetPasswordRequest) error {
	userID, err := s.repository.ConsumePasswordResetToken(ctx, tokenHash(request.Token))
	if err != nil {
		return ErrInvalidToken
	}
	return s.setPassword(ctx, userID, request.NewPassword, true)
}

func (s *Service) ChangePassword(ctx context.Context, userID string, request ChangePasswordRequest) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	return s.setPassword(ctx, userID, request.NewPassword, false)
}

func (s *Service) VerifyEmail(ctx context.Context, request VerifyEmailRequest) error {
	userID, err := s.repository.ConsumeEmailVerificationToken(ctx, tokenHash(request.Token))
	if err != nil {
		return ErrInvalidToken
	}
	return s.repository.MarkEmailVerified(ctx, userID)
}

func (s *Service) ParseToken(ctx context.Context, tokenString string) (Claims, error) {
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
	jti, _ := claims["jti"].(string)
	expiresUnix, _ := claims["exp"].(float64)
	expiresAt := time.Unix(int64(expiresUnix), 0)

	revoked, err := s.repository.IsAccessTokenRevoked(ctx, jti)
	if err != nil || revoked {
		return Claims{}, ErrInvalidToken
	}

	return Claims{
		UserID:         userID,
		Email:          email,
		Role:           role,
		OrganizationID: organizationID,
		JTI:            jti,
		ExpiresAt:      expiresAt,
	}, nil
}

func (s *Service) issueToken(user User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTokenTTL)
	organizationID := ""
	if user.OrganizationID != nil {
		organizationID = *user.OrganizationID
	}
	jti, err := randomToken()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := jwt.MapClaims{
		"sub":            user.ID,
		"email":          user.Email,
		"role":           user.PrimaryRole,
		"organizationId": organizationID,
		"jti":            jti,
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

func (s *Service) issueAuthTokens(ctx context.Context, user User, userAgent string, ipAddress string) (LoginResponse, error) {
	accessToken, expiresAt, err := s.issueToken(user)
	if err != nil {
		return LoginResponse{}, err
	}

	refreshToken, err := randomToken()
	if err != nil {
		return LoginResponse{}, err
	}
	if err := s.repository.CreateRefreshSession(ctx, user.ID, tokenHash(refreshToken), time.Now().Add(s.refreshTokenTTL), userAgent, ipAddress); err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		User:         user,
	}, nil
}

func (s *Service) setPassword(ctx context.Context, userID string, password string, revokeSessions bool) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repository.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return err
	}
	if revokeSessions {
		return s.repository.RevokeUserRefreshSessions(ctx, userID)
	}
	return nil
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type loginLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	attempts map[string]loginAttempt
}

type loginAttempt struct {
	count      int
	windowEnds time.Time
}

func newLoginLimiter(max int, window time.Duration) *loginLimiter {
	return &loginLimiter{
		max:      max,
		window:   window,
		attempts: map[string]loginAttempt{},
	}
}

func (l *loginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt := l.attempts[key]
	if time.Now().After(attempt.windowEnds) {
		return true
	}
	return attempt.count < l.max
}

func (l *loginLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	attempt := l.attempts[key]
	if now.After(attempt.windowEnds) {
		attempt = loginAttempt{windowEnds: now.Add(l.window)}
	}
	attempt.count++
	l.attempts[key] = attempt
}

func (l *loginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
