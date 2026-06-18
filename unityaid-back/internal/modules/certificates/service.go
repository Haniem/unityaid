package certificates

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

func (s *Service) List(ctx context.Context, userID string, all bool) ([]Certificate, error) {
	return s.repository.List(ctx, userID, all)
}

func (s *Service) Generate(ctx context.Context, request GenerateRequest, issuedBy string) (Certificate, error) {
	if strings.TrimSpace(request.Title) == "" {
		if request.Type == "hours" {
			request.Title = "Справка о волонтерских часах"
		} else {
			request.Title = "Сертификат участника"
		}
	}
	totalHours, err := s.repository.TotalHours(ctx, request.UserID, request.OrganizationID)
	if err != nil {
		return Certificate{}, err
	}
	return s.repository.Create(ctx, request, totalHours, verificationCode(), issuedBy)
}

func (s *Service) FindByID(ctx context.Context, id string) (Certificate, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) FindByCode(ctx context.Context, code string) (Certificate, error) {
	return s.repository.FindByCode(ctx, strings.TrimSpace(code))
}

func (s *Service) PDF(ctx context.Context, id string) ([]byte, Certificate, error) {
	item, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, Certificate{}, err
	}
	if item.Type == "participation" {
		item.Events, err = s.repository.ParticipationEvents(ctx, item.UserID, item.OrganizationID)
		if err != nil {
			return nil, Certificate{}, err
		}
	}
	return BuildPDF(item), item, nil
}

func verificationCode() string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "UA-CERT"
	}
	return "UA-" + strings.ToUpper(hex.EncodeToString(bytes))
}
