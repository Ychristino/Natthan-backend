package country

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	g := router.Group("/countries")
	g.Get("/", ctrl.List)
	g.Get("/:id", ctrl.GetByID)
	g.Post("/", authRequired, ctrl.Create)
	g.Patch("/:id", authRequired, ctrl.Update)
}
