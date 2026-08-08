package auth

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

func (c *Controller) Login(ctx *fiber.Ctx) error {
	var req LoginRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	resp, err := c.service.Login(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.JSON(resp)
}

func (c *Controller) Register(ctx *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	resp, err := c.service.Register(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

func (c *Controller) Refresh(ctx *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	resp, err := c.service.RefreshToken(ctx.Context(), req.RefreshToken)
	if err != nil {
		return err
	}
	return ctx.JSON(resp)
}

func (c *Controller) Logout(ctx *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	if err := c.service.Logout(ctx.Context(), req.RefreshToken); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) ListUsers(ctx *fiber.Ctx) error {
	var req ListUsersRequest
	if err := c.ParseQuery(ctx, &req); err != nil {
		return err
	}
	users, total, err := c.service.ListUsers(ctx.Context(), req)
	if err != nil {
		return err
	}
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(&u)
	}
	return ctx.JSON(UserListResponse{Data: resp, Total: total, Page: req.Page, Limit: req.Limit})
}

func (c *Controller) GetUser(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	u, err := c.service.GetUserByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(toUserResponse(u))
}

func (c *Controller) AddRole(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	var req AddRoleRequest
	if err := c.ParseBody(ctx, &req); err != nil {
		return err
	}
	ur, err := c.service.AddUserRole(ctx.Context(), id, req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(toUserRoleResponse(ur))
}

func (c *Controller) RemoveRole(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	role := ctx.Params("role")
	if role == "" {
		return fiber.NewError(fiber.StatusBadRequest, "role obrigatório")
	}
	if err := c.service.RemoveUserRole(ctx.Context(), id, role); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) DeactivateUser(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	if err := c.service.DeactivateUser(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) ReactivateUser(ctx *fiber.Ctx) error {
	id, err := c.ParseID(ctx)
	if err != nil {
		return err
	}
	if err := c.service.ReactivateUser(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

