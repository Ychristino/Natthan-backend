package serviceOrder

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	g := router.Group("/serviceOrders", authRequired)
	g.Get("/", ctrl.List)
	g.Post("/", ctrl.Create)
	g.Get("/:id", ctrl.GetByID)
	g.Patch("/:id", ctrl.Update)
	g.Delete("/:id", ctrl.Delete)

	g.Post("/:id/products", ctrl.AddProduct)
	g.Delete("/:id/products/:itemId", ctrl.RemoveProduct)
	g.Post("/:id/services", ctrl.AddService)
	g.Delete("/:id/services/:itemId", ctrl.RemoveService)
}
