package shop

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Wallet(ctx context.Context, userID string) (Wallet, error) {
	return s.repository.Wallet(ctx, userID)
}

func (s *Service) Transfer(ctx context.Context, userID string, req TransferRequest) (Wallet, error) {
	return s.repository.Transfer(ctx, userID, req)
}

func (s *Service) Products(ctx context.Context, organizationIDs []string) ([]Product, error) {
	return s.repository.Products(ctx, organizationIDs)
}

func (s *Service) CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error) {
	return s.repository.CreateProduct(ctx, req)
}

func (s *Service) CreateOrder(ctx context.Context, userID string, organizationIDs []string, req CreateOrderRequest) (Order, error) {
	return s.repository.CreateOrder(ctx, userID, organizationIDs, req)
}

func (s *Service) Orders(ctx context.Context, userID string, all bool, organizationIDs []string) ([]Order, error) {
	return s.repository.Orders(ctx, userID, all, organizationIDs)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id string, status string, managerID string) (Order, error) {
	return s.repository.UpdateOrderStatus(ctx, id, status, managerID)
}

func (s *Service) Order(ctx context.Context, id string) (Order, error) {
	return s.repository.Order(ctx, id)
}
