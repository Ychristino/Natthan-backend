package stock

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type Service interface {
	List(ctx context.Context, req ListStockRequest) ([]StockResponse, int64, error)
	ListMovements(ctx context.Context, req ListMovementsRequest) ([]MovementResponse, int64, error)
	CreateMovement(ctx context.Context, req CreateMovementRequest, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, req ListStockRequest) ([]StockResponse, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	items, total, err := s.repo.List(ctx, req.Name, page.Limit, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]StockResponse, len(items))
	for i, item := range items {
		resp[i] = toStockResponse(item)
	}
	return resp, total, nil
}

func (s *service) ListMovements(ctx context.Context, req ListMovementsRequest) ([]MovementResponse, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	movements, total, err := s.repo.ListMovements(ctx, req.ProductID, req.Type, page.Limit, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]MovementResponse, len(movements))
	for i, m := range movements {
		resp[i] = toMovementResponse(m)
	}
	return resp, total, nil
}

func (s *service) CreateMovement(ctx context.Context, req CreateMovementRequest, userID uuid.UUID) error {
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "product_id inválido")
	}
	if err := s.repo.ApplyMovement(ctx, productID, MovementType(req.Type), req.Quantity, userID); err != nil {
		if errors.Is(err, ErrInsufficientStock) {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "estoque insuficiente")
		}
		return err
	}
	return nil
}
