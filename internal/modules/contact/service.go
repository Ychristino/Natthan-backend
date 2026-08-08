package contact

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Contact, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Contact, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Contact, error)
	List(ctx context.Context, req ListRequest) ([]Contact, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Contact, error) {
	personID, _ := uuid.Parse(req.PersonID)
	return s.repo.Create(ctx, Contact{
		ID:       uuid.New(),
		PersonID: personID,
		Type:     req.Type,
		Value:    req.Value,
	})
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Contact, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "contato não encontrado")
		}
		return nil, err
	}
	return c, nil
}

func (s *service) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Contact, error) {
	return s.repo.GetByPersonID(ctx, personID)
}

func (s *service) List(ctx context.Context, req ListRequest) ([]Contact, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	var personID *uuid.UUID
	if req.PersonID != "" {
		id, _ := uuid.Parse(req.PersonID)
		personID = &id
	}
	return s.repo.List(ctx, personID, page.Limit, page.Offset)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "contato não encontrado")
	}
	return err
}
