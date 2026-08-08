package serviceOrder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/base"
)

func parseLocalUUID(ctx *fiber.Ctx, key string) (uuid.UUID, error) {
	s, _ := ctx.Locals(key).(string)
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "autenticação inválida")
	}
	return id, nil
}

type Controller struct {
	base.Controller
	service Service
}

func NewController(svc Service) *Controller {
	return &Controller{Controller: base.NewController(), service: svc}
}

func (c *Controller) List(ctx *fiber.Ctx) error {
	var req ListRequest
	if err := c.ParseQuery(ctx, &req); err != nil {
		return err
	}
	orders, total, err := c.service.List(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.JSON(ListResponse{Data: orders, Total: total, Page: req.Page, Limit: req.Limit})
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	so, err := c.service.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(so)
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	employeePersonID, err := parseLocalUUID(ctx, "person_id")
	if err != nil {
		return err
	}
	userID, err := parseLocalUUID(ctx, "user_id")
	if err != nil {
		return err
	}

	var req CreateRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	so, err := c.service.Create(ctx.Context(), req, employeePersonID, userID)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(so) // DetailResponse with items
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	var req UpdateRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	so, err := c.service.Update(ctx.Context(), id, req)
	if err != nil {
		return err
	}
	return ctx.JSON(so)
}

func (c *Controller) Delete(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) AddProduct(ctx *fiber.Ctx) error {
	orderID, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	userID, err := parseLocalUUID(ctx, "user_id")
	if err != nil {
		return err
	}
	var req AddProductRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	item, err := c.service.AddProduct(ctx.Context(), orderID, req, userID)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(item)
}

func (c *Controller) RemoveProduct(ctx *fiber.Ctx) error {
	orderID, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	itemID, err := uuid.Parse(ctx.Params("itemId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "itemId inválido")
	}
	if err := c.service.RemoveProduct(ctx.Context(), orderID, itemID); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) AddService(ctx *fiber.Ctx) error {
	orderID, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	var req AddServiceRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	item, err := c.service.AddService(ctx.Context(), orderID, req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(item)
}

func (c *Controller) RemoveService(ctx *fiber.Ctx) error {
	orderID, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	itemID, err := uuid.Parse(ctx.Params("itemId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "itemId inválido")
	}
	if err := c.service.RemoveService(ctx.Context(), orderID, itemID); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
