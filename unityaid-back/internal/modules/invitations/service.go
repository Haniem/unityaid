package invitations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Invitation, error) {
	return s.repository.List(ctx)
}

func (s *Service) Create(ctx context.Context, request CreateRequest, invitedBy string) (Invitation, error) {
	role := strings.TrimSpace(request.Role)
	if role == "" {
		role = "volunteer"
	}
	token, err := newToken()
	if err != nil {
		return Invitation{}, err
	}
	return s.repository.Create(ctx, strings.ToLower(strings.TrimSpace(request.Email)), role, token, invitedBy)
}

func (s *Service) Accept(ctx context.Context, token string, userID string) (Invitation, error) {
	return s.repository.Accept(ctx, strings.TrimSpace(token), userID)
}

func newToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
