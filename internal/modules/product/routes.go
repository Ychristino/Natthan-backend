package product

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	products := router.Group("/products")
	products.Use(authRequired)

	products.Get("/", ctrl.List)        // GET    /api/v1/products
	products.Post("/", ctrl.Create)     // POST   /api/v1/products
	products.Get("/:id", ctrl.GetByID)  // GET    /api/v1/products/:id
	products.Patch("/:id", ctrl.Update) // PATCH  /api/v1/products/:id
	products.Delete("/:id", ctrl.Delete) // DELETE /api/v1/products/:id
}
