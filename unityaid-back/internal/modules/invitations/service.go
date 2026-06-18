package invitations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
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
	if role != "volunteer" && role != "coordinator" && role != "org_admin" {
		role = "volunteer"
	}
	token, err := newToken()
	if err != nil {
		return Invitation{}, err
	}
	var email *string
	if normalizedEmail := strings.ToLower(strings.TrimSpace(request.Email)); normalizedEmail != "" {
		email = &normalizedEmail
	}
	var organizationID *string
	if normalizedOrganizationID := strings.TrimSpace(request.OrganizationID); normalizedOrganizationID != "" {
		organizationID = &normalizedOrganizationID
	}
	return s.repository.Create(ctx, email, role, token, invitedBy, organizationID)
}

func (s *Service) FindByToken(ctx context.Context, token string) (Invitation, error) {
	return s.repository.FindByToken(ctx, strings.TrimSpace(token))
}

func (s *Service) AcceptRegistration(ctx context.Context, token string, request AcceptRegistrationRequest) (Invitation, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return Invitation{}, err
	}
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Patronymic = strings.TrimSpace(request.Patronymic)
	return s.repository.AcceptRegistration(ctx, strings.TrimSpace(token), request, string(passwordHash))
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
