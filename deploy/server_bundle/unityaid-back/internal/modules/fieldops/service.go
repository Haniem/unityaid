package fieldops

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateQRToken(ctx context.Context, eventID string, request CreateQRRequest, createdBy string) (QRToken, error) {
	return s.repository.CreateQRToken(ctx, eventID, request, createdBy)
}

func (s *Service) Scan(ctx context.Context, token string, userID string) (Checkin, error) {
	return s.repository.Scan(ctx, token, userID)
}
