package address

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
	"github.com/natthan/api/internal/modules/city"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*AddressDetail, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AddressDetail, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) ([]AddressDetail, error)
	List(ctx context.Context, req ListRequest) ([]AddressDetail, int64, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*AddressDetail, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo    Repository
	citySvc city.Service
}

func NewService(repo Repository, citySvc city.Service) Service {
	return &service{repo: repo, citySvc: citySvc}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*AddressDetail, error) {
	cityID, _ := uuid.Parse(req.CityID)
	if _, err := s.citySvc.GetByID(ctx, cityID); err != nil {
		return nil, err
	}
	personID, _ := uuid.Parse(req.PersonID)

	a := Address{
		ID:           uuid.New(),
		PersonID:     personID,
		AddressType:  req.AddressType,
		Street:       req.Street,
		Number:       req.Number,
		Complement:   req.Complement,
		Neighborhood: req.Neighborhood,
		ZipCode:      req.ZipCode,
		CityID:       cityID,
	}

	return s.repo.Create(ctx, a)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*AddressDetail, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "endereço não encontrado")
		}
		return nil, err
	}
	return d, nil
}

func (s *service) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]AddressDetail, error) {
	return s.repo.GetByPersonID(ctx, personID)
}

func (s *service) List(ctx context.Context, req ListRequest) ([]AddressDetail, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	var personID *uuid.UUID
	if req.PersonID != "" {
		id, _ := uuid.Parse(req.PersonID)
		personID = &id
	}
	return s.repo.List(ctx, personID, page.Limit, page.Offset)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*AddressDetail, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "endereço não encontrado")
		}
		return nil, err
	}

	a := Address{
		ID:           id,
		PersonID:     existing.PersonID,
		AddressType:  existing.AddressType,
		Street:       existing.Street,
		Number:       existing.Number,
		Complement:   existing.Complement,
		Neighborhood: existing.Neighborhood,
		ZipCode:      existing.ZipCode,
		CityID:       existing.CityID,
		UpdatedAt:    time.Now(),
	}

	if req.AddressType != nil {
		a.AddressType = *req.AddressType
	}
	if req.Street != nil {
		a.Street = *req.Street
	}
	if req.Number != nil {
		a.Number = *req.Number
	}
	if req.Complement != nil {
		a.Complement = *req.Complement
	}
	if req.Neighborhood != nil {
		a.Neighborhood = *req.Neighborhood
	}
	if req.ZipCode != nil {
		a.ZipCode = *req.ZipCode
	}
	if req.CityID != nil {
		cityID, _ := uuid.Parse(*req.CityID)
		if _, err := s.citySvc.GetByID(ctx, cityID); err != nil {
			return nil, err
		}
		a.CityID = cityID
	}

	return s.repo.Update(ctx, a)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "endereço não encontrado")
	}
	return err
}
