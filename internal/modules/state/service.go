package state

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
	"github.com/natthan/api/internal/modules/country"
)

type Service interface {
	List(ctx context.Context, req ListRequest) ([]State, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*State, error)
	Create(ctx context.Context, req CreateRequest) (*State, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*State, error)
}

type service struct {
	repo        Repository
	countrySvc  country.Service
}

func NewService(repo Repository, countrySvc country.Service) Service {
	return &service{repo: repo, countrySvc: countrySvc}
}

func (s *service) List(ctx context.Context, req ListRequest) ([]State, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)

	var countryID *uuid.UUID
	if req.CountryID != "" {
		id, err := uuid.Parse(req.CountryID)
		if err != nil {
			return nil, 0, fiber.NewError(fiber.StatusBadRequest, "country_id inválido")
		}
		countryID = &id
	}

	return s.repo.List(ctx, req.Name, countryID, page.Limit, page.Offset)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*State, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "estado não encontrado")
		}
		return nil, err
	}
	return st, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*State, error) {
	countryID, _ := uuid.Parse(req.CountryID)
	if _, err := s.countrySvc.GetByID(ctx, countryID); err != nil {
		return nil, err
	}

	st := State{ID: uuid.New(), Name: req.Name, Code: req.Code, CountryID: countryID}
	return s.repo.Create(ctx, st)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*State, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "estado não encontrado")
		}
		return nil, err
	}
	if req.Name != nil {
		st.Name = *req.Name
	}
	if req.Code != nil {
		st.Code = *req.Code
	}
	return s.repo.Update(ctx, *st)
}
