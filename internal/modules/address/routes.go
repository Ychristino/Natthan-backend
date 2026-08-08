package address

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	g := router.Group("/addresses", authRequired)
	g.Get("/", ctrl.List)
	g.Get("/:id", ctrl.GetByID)
	g.Post("/", ctrl.Create)
	g.Patch("/:id", ctrl.Update)
	g.Delete("/:id", ctrl.Delete)
}
