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

func (s *Service) Products(ctx context.Context) ([]Product, error) {
	return s.repository.Products(ctx)
}

func (s *Service) CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error) {
	return s.repository.CreateProduct(ctx, req)
}

func (s *Service) CreateOrder(ctx context.Context, userID string, req CreateOrderRequest) (Order, error) {
	return s.repository.CreateOrder(ctx, userID, req)
}

func (s *Service) Orders(ctx context.Context, userID string, all bool) ([]Order, error) {
	return s.repository.Orders(ctx, userID, all)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id string, status string, managerID string) (Order, error) {
	return s.repository.UpdateOrderStatus(ctx, id, status, managerID)
}
