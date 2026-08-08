package country

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type Service interface {
	List(ctx context.Context, req ListRequest) ([]Country, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Country, error)
	Create(ctx context.Context, req CreateRequest) (*Country, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Country, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, req ListRequest) ([]Country, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	return s.repo.List(ctx, req.Name, page.Limit, page.Offset)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Country, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "país não encontrado")
		}
		return nil, err
	}
	return c, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Country, error) {
	c := Country{ID: uuid.New(), Name: req.Name, Code: req.Code}
	return s.repo.Create(ctx, c)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Country, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "país não encontrado")
		}
		return nil, err
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Code != nil {
		c.Code = *req.Code
	}
	return s.repo.Update(ctx, *c)
}
