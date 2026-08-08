package city

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
	"github.com/natthan/api/internal/modules/state"
)

type Service interface {
	List(ctx context.Context, req ListRequest) ([]City, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*City, error)
	Create(ctx context.Context, req CreateRequest) (*City, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*City, error)
}

type service struct {
	repo     Repository
	stateSvc state.Service
}

func NewService(repo Repository, stateSvc state.Service) Service {
	return &service{repo: repo, stateSvc: stateSvc}
}

func (s *service) List(ctx context.Context, req ListRequest) ([]City, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)

	var stateID *uuid.UUID
	if req.StateID != "" {
		id, err := uuid.Parse(req.StateID)
		if err != nil {
			return nil, 0, fiber.NewError(fiber.StatusBadRequest, "state_id inválido")
		}
		stateID = &id
	}

	return s.repo.List(ctx, req.Name, stateID, page.Limit, page.Offset)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*City, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "cidade não encontrada")
		}
		return nil, err
	}
	return c, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*City, error) {
	stateID, _ := uuid.Parse(req.StateID)
	if _, err := s.stateSvc.GetByID(ctx, stateID); err != nil {
		return nil, err
	}

	c := City{ID: uuid.New(), Name: req.Name, StateID: stateID}
	return s.repo.Create(ctx, c)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*City, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "cidade não encontrada")
		}
		return nil, err
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	return s.repo.Update(ctx, *c)
}
