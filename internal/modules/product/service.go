package product

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Product, error)
	List(ctx context.Context, req ListRequest) ([]Product, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type productService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, req CreateRequest) (*Product, error) {
	p := Product{
		ID:            uuid.New(),
		ProductCode:   req.ProductCode,
		Name:          req.Name,
		Description:   req.Description,
		PurchasePrice: req.PurchasePrice,
		SalePrice:     req.SalePrice,
	}
	return s.repo.Create(ctx, p)
}

func (s *productService) List(ctx context.Context, req ListRequest) ([]Product, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	return s.repo.List(ctx, req.Name, page.Limit, page.Offset)
}

func (s *productService) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "produto não encontrado")
		}
		return nil, err
	}
	return p, nil
}

func (s *productService) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "produto não encontrado")
		}
		return nil, err
	}
	if req.ProductCode != nil {
		p.ProductCode = *req.ProductCode
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.PurchasePrice != nil {
		p.PurchasePrice = *req.PurchasePrice
	}
	if req.SalePrice != nil {
		p.SalePrice = *req.SalePrice
	}
	return s.repo.Update(ctx, *p)
}

func (s *productService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "produto não encontrado")
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}
