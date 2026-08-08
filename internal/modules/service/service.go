package service

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type ServiceInterface interface {
	Create(ctx context.Context, req CreateRequest) (*Service, error)
	List(ctx context.Context, req ListRequest) ([]Service, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Service, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Service, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type svc struct {
	repo Repository
}

func NewService(repo Repository) ServiceInterface {
	return &svc{repo: repo}
}

func (s *svc) Create(ctx context.Context, req CreateRequest) (*Service, error) {
	service := Service{
		ID:          uuid.New(),
		ServiceCode: req.ServiceCode,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}
	return s.repo.Create(ctx, service)
}

func (s *svc) List(ctx context.Context, req ListRequest) ([]Service, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	return s.repo.List(ctx, req.Name, page.Limit, page.Offset)
}

func (s *svc) GetByID(ctx context.Context, id uuid.UUID) (*Service, error) {
	sv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "serviço não encontrado")
		}
		return nil, err
	}
	return sv, nil
}

func (s *svc) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Service, error) {
	sv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "serviço não encontrado")
		}
		return nil, err
	}
	if req.ServiceCode != nil {
		sv.ServiceCode = *req.ServiceCode
	}
	if req.Name != nil {
		sv.Name = *req.Name
	}
	if req.Description != nil {
		sv.Description = *req.Description
	}
	if req.Price != nil {
		sv.Price = *req.Price
	}
	return s.repo.Update(ctx, *sv)
}

func (s *svc) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "serviço não encontrado")
		}
		return err
	}
	return s.repo.Deactivate(ctx, id)
}
