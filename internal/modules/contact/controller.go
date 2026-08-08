package contact

import (
	"github.com/gofiber/fiber/v2"
	"github.com/natthan/api/internal/core/base"
)

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
	contacts, total, err := c.service.List(ctx.Context(), req)
	if err != nil {
		return err
	}
	resp := make([]Response, len(contacts))
	for i, ct := range contacts {
		resp[i] = ToResponse(ct)
	}
	return ctx.JSON(ListResponse{Data: resp, Total: total, Page: req.Page, Limit: req.Limit})
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	ct, err := c.service.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(ToResponse(*ct))
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	var req CreateRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	ct, err := c.service.Create(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(ToResponse(*ct))
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
