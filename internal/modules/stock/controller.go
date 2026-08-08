package stock

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	var req ListStockRequest
	if err := c.ParseQuery(ctx, &req); err != nil {
		return err
	}
	items, total, err := c.service.List(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.JSON(StockListResponse{Data: items, Total: total, Page: req.Page, Limit: req.Limit})
}

func (c *Controller) ListMovements(ctx *fiber.Ctx) error {
	var req ListMovementsRequest
	if err := c.ParseQuery(ctx, &req); err != nil {
		return err
	}
	movements, total, err := c.service.ListMovements(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.JSON(MovementListResponse{Data: movements, Total: total, Page: req.Page, Limit: req.Limit})
}

func (c *Controller) CreateMovement(ctx *fiber.Ctx) error {
	userID, err := parseLocalUUID(ctx, "user_id")
	if err != nil {
		return err
	}
	var req CreateMovementRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	if err := c.service.CreateMovement(ctx.Context(), req, userID); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func parseLocalUUID(ctx *fiber.Ctx, key string) (uuid.UUID, error) {
	s, _ := ctx.Locals(key).(string)
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "autenticação inválida")
	}
	return id, nil
}
