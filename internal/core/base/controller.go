package base

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Controller is embedded into every module controller to provide
// shared parsing, validation, and ID extraction without repetition.
type Controller struct {
	validate *validator.Validate
}

func NewController() Controller {
	return Controller{validate: validator.New()}
}

// ParseBody parses the request body into req and validates it.
func (b *Controller) ParseBody(ctx *fiber.Ctx, req any) error {
	if err := ctx.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "corpo da requisição inválido")
	}
	if err := b.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	return nil
}

// ParseQuery parses query string parameters into req and validates it.
func (b *Controller) ParseQuery(ctx *fiber.Ctx, req any) error {
	if err := ctx.QueryParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "parâmetros inválidos")
	}
	if err := b.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	return nil
}

// ParseID extracts and parses a UUID from the route :id parameter.
func (b *Controller) ParseID(ctx *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return uuid.UUID{}, fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	return id, nil
}
