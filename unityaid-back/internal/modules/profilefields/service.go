package profilefields

import (
	"context"
	"errors"
	"strings"
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) Schema(ctx context.Context, organizationID string) ([]Group, error) {
	return s.repository.ListSchema(ctx, strings.TrimSpace(organizationID))
}

func (s *Service) CreateGroup(ctx context.Context, request GroupRequest) (Group, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	if !request.IsActive {
		request.IsActive = true
	}
	return s.repository.CreateGroup(ctx, request)
}
func (s *Service) UpdateGroup(ctx context.Context, id string, request GroupRequest) (Group, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	return s.repository.UpdateGroup(ctx, id, request)
}
func (s *Service) DeleteGroup(ctx context.Context, id, org string) error {
	return s.repository.DeleteGroup(ctx, id, org)
}

func (s *Service) CreateField(ctx context.Context, request FieldRequest) (Field, error) {
	if !validType(request.Type) {
		return Field{}, errors.New("unsupported field type")
	}
	request.Name = strings.TrimSpace(request.Name)
	return s.repository.CreateField(ctx, request)
}
func (s *Service) UpdateField(ctx context.Context, id string, request FieldRequest) (Field, error) {
	if !validType(request.Type) {
		return Field{}, errors.New("unsupported field type")
	}
	request.Name = strings.TrimSpace(request.Name)
	return s.repository.UpdateField(ctx, id, request)
}
func (s *Service) DeleteField(ctx context.Context, id, org string) error {
	return s.repository.DeleteField(ctx, id, org)
}
func (s *Service) Values(ctx context.Context, org, user string) (ValuesResponse, error) {
	values, err := s.repository.Values(ctx, org, user)
	return ValuesResponse{Values: values}, err
}
func (s *Service) SaveValues(ctx context.Context, user, editor string, request ValuesRequest) (ValuesResponse, error) {
	if err := s.repository.SaveValues(ctx, request.OrganizationID, user, editor, request.Values); err != nil {
		return ValuesResponse{}, err
	}
	return s.Values(ctx, request.OrganizationID, user)
}
func validType(value string) bool {
	for _, item := range []string{"text", "textarea", "number", "date", "datetime", "tel", "email", "url", "select", "multiselect", "checkbox", "file"} {
		if item == value {
			return true
		}
	}
	return false
}
