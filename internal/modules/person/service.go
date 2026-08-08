package person

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
	"github.com/natthan/api/internal/modules/address"
	"github.com/natthan/api/internal/modules/contact"
	"github.com/natthan/api/internal/modules/document"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*PersonResponse, error)
	List(ctx context.Context, req ListRequest) ([]PersonResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*PersonResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*PersonResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	LookupByDocument(ctx context.Context, value string) (*LookupResult, error)
}

type service struct {
	repo        Repository
	addressSvc  address.Service
	contactSvc  contact.Service
	documentSvc document.Service
}

func NewService(repo Repository, addressSvc address.Service, contactSvc contact.Service, documentSvc document.Service) Service {
	return &service{repo: repo, addressSvc: addressSvc, contactSvc: contactSvc, documentSvc: documentSvc}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*PersonResponse, error) {
	p := Person{
		ID:       uuid.New(),
		FullName: req.FullName,
	}
	if req.BirthDate != nil {
		t, _ := time.Parse("2006-01-02", *req.BirthDate)
		p.BirthDate = &t
	}
	if req.NationalityID != nil {
		id, _ := uuid.Parse(*req.NationalityID)
		p.NationalityID = &id
	}

	created, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	for _, dr := range req.Documents {
		if _, err := s.documentSvc.Create(ctx, document.CreateRequest{
			PersonID: created.ID.String(),
			Type:     dr.Type,
			Value:    dr.Value,
		}); err != nil {
			return nil, err
		}
	}

	for _, ar := range req.Addresses {
		if _, err := s.addressSvc.Create(ctx, address.CreateRequest{
			PersonID:     created.ID.String(),
			AddressType:  ar.AddressType,
			Street:       ar.Street,
			Number:       ar.Number,
			Complement:   ar.Complement,
			Neighborhood: ar.Neighborhood,
			ZipCode:      ar.ZipCode,
			CityID:       ar.CityID,
		}); err != nil {
			return nil, err
		}
	}

	for _, cr := range req.Contacts {
		if _, err := s.contactSvc.Create(ctx, contact.CreateRequest{
			PersonID: created.ID.String(),
			Type:     cr.Type,
			Value:    cr.Value,
		}); err != nil {
			return nil, err
		}
	}

	return s.GetByID(ctx, created.ID)
}

func (s *service) List(ctx context.Context, req ListRequest) ([]PersonResponse, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	persons, total, err := s.repo.List(ctx, req.Name, page.Limit, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]PersonResponse, len(persons))
	for i, p := range persons {
		resp[i] = toResponse(p)
	}
	return resp, total, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*PersonResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "pessoa não encontrada")
		}
		return nil, err
	}

	var docs []document.Document
	var addrs []address.AddressDetail
	var contacts []contact.Contact

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var e error
		docs, e = s.documentSvc.GetByPersonID(gctx, p.ID)
		return e
	})

	g.Go(func() error {
		var e error
		addrs, e = s.addressSvc.GetByPersonID(gctx, p.ID)
		return e
	})

	g.Go(func() error {
		var e error
		contacts, e = s.contactSvc.GetByPersonID(gctx, p.ID)
		return e
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	resp := toResponse(*p)
	resp.Documents = toDocumentResponses(docs)
	resp.Addresses = toAddressResponses(addrs)
	resp.Contacts = toContactResponses(contacts)

	return &resp, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*PersonResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "pessoa não encontrada")
		}
		return nil, err
	}

	if req.FullName != nil {
		p.FullName = *req.FullName
	}
	if req.BirthDate != nil {
		t, _ := time.Parse("2006-01-02", *req.BirthDate)
		p.BirthDate = &t
	}
	if req.NationalityID != nil {
		id, _ := uuid.Parse(*req.NationalityID)
		p.NationalityID = &id
	}

	updated, err := s.repo.Update(ctx, *p)
	if err != nil {
		return nil, err
	}

	resp := toResponse(*updated)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "pessoa não encontrada")
	}
	return err
}

func (s *service) LookupByDocument(ctx context.Context, value string) (*LookupResult, error) {
	p, err := s.repo.GetPersonByDocumentValue(ctx, value)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return &LookupResult{Found: false}, nil
	}
	return &LookupResult{Found: true, PersonID: p.ID, FullName: p.FullName}, nil
}

