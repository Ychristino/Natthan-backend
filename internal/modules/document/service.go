package document

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Document, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Document, error)
	List(ctx context.Context, req ListRequest) ([]Document, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Document, error) {
	personID, _ := uuid.Parse(req.PersonID)
	return s.repo.Create(ctx, Document{
		ID:       uuid.New(),
		PersonID: personID,
		Type:     req.Type,
		Value:    req.Value,
	})
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "documento não encontrado")
		}
		return nil, err
	}
	return d, nil
}

func (s *service) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Document, error) {
	return s.repo.GetByPersonID(ctx, personID)
}

func (s *service) List(ctx context.Context, req ListRequest) ([]Document, int64, error) {
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
		return fiber.NewError(fiber.StatusNotFound, "documento não encontrado")
	}
	return err
}
