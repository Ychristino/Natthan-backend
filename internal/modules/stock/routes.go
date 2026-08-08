package stock

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	s := router.Group("/stock")
	s.Use(authRequired)

	s.Get("/", ctrl.List)                    // GET  /api/v1/stock
	s.Get("/movements", ctrl.ListMovements)  // GET  /api/v1/stock/movements
	s.Post("/movements", ctrl.CreateMovement) // POST /api/v1/stock/movements
}
