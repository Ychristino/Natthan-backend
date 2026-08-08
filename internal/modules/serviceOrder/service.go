package serviceOrder

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/pagination"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest, employeePersonID, userID uuid.UUID) (*DetailResponse, error)
	List(ctx context.Context, req ListRequest) ([]Response, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*DetailResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Response, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddProduct(ctx context.Context, orderID uuid.UUID, req AddProductRequest, userID uuid.UUID) (*OrderProductResponse, error)
	RemoveProduct(ctx context.Context, orderID, itemID uuid.UUID) error
	AddService(ctx context.Context, orderID uuid.UUID, req AddServiceRequest) (*OrderServiceResponse, error)
	RemoveService(ctx context.Context, orderID, itemID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest, employeePersonID, userID uuid.UUID) (*DetailResponse, error) {
	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "client_id inválido")
	}

	so := ServiceOrder{
		ID:         uuid.New(),
		ClientID:   clientID,
		EmployeeID: employeePersonID,
		Status:     StatusOpen,
		EntryDate:  time.Now(),
		IsPaid:     req.IsPaid,
		Notes:      req.Notes,
	}
	if req.Status != "" {
		so.Status = Status(req.Status)
	}
	if req.PayDate != nil {
		t, err := time.Parse("2006-01-02", *req.PayDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "pay_date inválida")
		}
		so.PayDate = &t
	}

	// Validate and build product/service slices before touching the DB.
	products := make([]OrderProduct, 0, len(req.Products))
	for _, p := range req.Products {
		productID, err := uuid.Parse(p.ProductID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "product_id inválido")
		}
		products = append(products, OrderProduct{
			ID:        uuid.New(),
			ProductID: productID,
			Quantity:  p.Quantity,
			UnitPrice: p.UnitPrice,
		})
	}

	services := make([]OrderService, 0, len(req.Services))
	for _, sv := range req.Services {
		serviceID, err := uuid.Parse(sv.ServiceID)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "service_id inválido")
		}
		services = append(services, OrderService{
			ID:           uuid.New(),
			ServiceID:    serviceID,
			PriceApplied: sv.PriceApplied,
			Notes:        sv.Notes,
		})
	}

	// Single atomic transaction: order + stock deductions + movements + line items.
	full, err := s.repo.CreateFull(ctx, so, products, services, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrInsufficientStock):
			return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "estoque insuficiente para um dos produtos")
		case errors.Is(err, ErrClientNotFound):
			return nil, fiber.NewError(fiber.StatusBadRequest, "cliente não encontrado")
		case errors.Is(err, ErrProductNotFound):
			return nil, fiber.NewError(fiber.StatusBadRequest, "produto não encontrado")
		case errors.Is(err, ErrServiceNotFound):
			return nil, fiber.NewError(fiber.StatusBadRequest, "serviço não encontrado")
		}
		return nil, err
	}
	r := toDetailResponse(*full)
	return &r, nil
}

func (s *service) List(ctx context.Context, req ListRequest) ([]Response, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	orders, total, err := s.repo.List(ctx, req.Search, req.Status, req.DateFrom, req.DateTo, page.Limit, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]Response, len(orders))
	for i, o := range orders {
		resp[i] = toResponse(o)
	}
	return resp, total, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*DetailResponse, error) {
	so, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "ordem de serviço não encontrada")
		}
		return nil, err
	}
	r := toDetailResponse(*so)
	return &r, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Response, error) {
	so, err := s.repo.GetBase(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "ordem de serviço não encontrada")
		}
		return nil, err
	}
	if req.Status != nil {
		so.Status = Status(*req.Status)
	}
	if req.ExitDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ExitDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "exit_date inválida")
		}
		so.ExitDate = &t
	}
	if req.PayDate != nil {
		t, err := time.Parse("2006-01-02", *req.PayDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "pay_date inválida")
		}
		so.PayDate = &t
	}
	if req.IsPaid != nil {
		so.IsPaid = *req.IsPaid
	}
	if req.Notes != nil {
		so.Notes = *req.Notes
	}
	updated, err := s.repo.Update(ctx, *so)
	if err != nil {
		return nil, err
	}
	r := toResponse(*updated)
	return &r, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Deactivate(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "ordem de serviço não encontrada")
	}
	return err
}

func (s *service) AddProduct(ctx context.Context, orderID uuid.UUID, req AddProductRequest, userID uuid.UUID) (*OrderProductResponse, error) {
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "product_id inválido")
	}
	item := OrderProduct{
		ID:             uuid.New(),
		ServiceOrderID: orderID,
		ProductID:      productID,
		Quantity:       req.Quantity,
		UnitPrice:      req.UnitPrice,
	}
	p, err := s.repo.AddProduct(ctx, item, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrInsufficientStock):
			return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "estoque insuficiente")
		case errors.Is(err, ErrProductNotFound):
			return nil, fiber.NewError(fiber.StatusBadRequest, "produto não encontrado")
		}
		return nil, err
	}
	r := toOrderProductResponse(*p)
	return &r, nil
}

func (s *service) RemoveProduct(ctx context.Context, orderID, itemID uuid.UUID) error {
	err := s.repo.RemoveProduct(ctx, orderID, itemID)
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "item não encontrado nesta ordem")
	}
	return err
}

func (s *service) AddService(ctx context.Context, orderID uuid.UUID, req AddServiceRequest) (*OrderServiceResponse, error) {
	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "service_id inválido")
	}
	item := OrderService{
		ID:             uuid.New(),
		ServiceOrderID: orderID,
		ServiceID:      serviceID,
		PriceApplied:   req.PriceApplied,
		Notes:          req.Notes,
	}
	sv, err := s.repo.AddService(ctx, item)
	if err != nil {
		if errors.Is(err, ErrServiceNotFound) {
			return nil, fiber.NewError(fiber.StatusBadRequest, "serviço não encontrado")
		}
		return nil, err
	}
	r := toOrderServiceResponse(*sv)
	return &r, nil
}

func (s *service) RemoveService(ctx context.Context, orderID, itemID uuid.UUID) error {
	err := s.repo.RemoveService(ctx, orderID, itemID)
	if errors.Is(err, ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "item não encontrado nesta ordem")
	}
	return err
}

