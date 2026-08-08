package person

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
	persons, total, err := c.service.List(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.JSON(ListResponse{Data: persons, Total: total, Page: req.Page, Limit: req.Limit})
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	p, err := c.service.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(p)
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	var req CreateRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	p, err := c.service.Create(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(p)
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
	p, err := c.service.Update(ctx.Context(), id, req)
	if err != nil {
		return err
	}
	return ctx.JSON(p)
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

func (c *Controller) Lookup(ctx *fiber.Ctx) error {
	doc := ctx.Query("document")
	if doc == "" {
		return fiber.NewError(fiber.StatusBadRequest, "parâmetro 'document' obrigatório")
	}
	result, err := c.service.LookupByDocument(ctx.Context(), doc)
	if err != nil {
		return err
	}
	resp := LookupResponse{Found: result.Found}
	if result.Found {
		resp.PersonID = result.PersonID.String()
		resp.FullName = result.FullName
	}
	return ctx.JSON(resp)
}
