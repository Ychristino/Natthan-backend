package contact

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	g := router.Group("/contacts", authRequired)
	g.Get("/", ctrl.List)
	g.Get("/:id", ctrl.GetByID)
	g.Post("/", ctrl.Create)
	g.Delete("/:id", ctrl.Delete)
}
