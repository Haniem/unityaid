package timeentries

import (
	"context"

	"unityaid-back/internal/modules/auth"
)

type Service struct {
	repository *Repository
	authorizer *auth.Authorizer
}

func NewService(repository *Repository, authorizer *auth.Authorizer) *Service {
	return &Service{repository: repository, authorizer: authorizer}
}

func (s *Service) List(ctx context.Context, claims auth.Claims, filters ListFilters) ([]TimeEntry, error) {
	organizationIDs, err := s.authorizer.ManageableOrganizationIDs(ctx, claims)
	if err != nil {
		return nil, err
	}
	if organizationIDs == nil {
		return s.repository.List(ctx, filters)
	}
	if len(organizationIDs) > 0 {
		return s.repository.ListForOrganizations(ctx, organizationIDs, filters)
	}
	filters.UserID = claims.UserID
	return s.repository.List(ctx, filters)
}

func (s *Service) FindByID(ctx context.Context, id string) (TimeEntry, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, userID string, request UpsertRequest) (TimeEntry, error) {
	return s.repository.Create(ctx, userID, request)
}

func (s *Service) ResolveOrganization(ctx context.Context, request UpsertRequest) (string, error) {
	return s.repository.ResolveOrganization(ctx, request.EventID, request.TaskID)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest) (TimeEntry, error) {
	return s.repository.Update(ctx, id, request)
}

func (s *Service) Approve(ctx context.Context, id string, reviewerID string) (TimeEntry, error) {
	return s.repository.Approve(ctx, id, reviewerID)
}

func (s *Service) Reject(ctx context.Context, id string, reviewerID string) (TimeEntry, error) {
	return s.repository.Reject(ctx, id, reviewerID)
}
