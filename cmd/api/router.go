package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/natthan/api/internal/middleware"
	"github.com/natthan/api/internal/modules/address"
	"github.com/natthan/api/internal/modules/auth"
	"github.com/natthan/api/internal/modules/city"
	"github.com/natthan/api/internal/modules/contact"
	"github.com/natthan/api/internal/modules/country"
	"github.com/natthan/api/internal/modules/document"
	"github.com/natthan/api/internal/modules/person"
	"github.com/natthan/api/internal/modules/product"
	"github.com/natthan/api/internal/modules/service"
	"github.com/natthan/api/internal/modules/serviceOrder"
	"github.com/natthan/api/internal/modules/state"
	"github.com/natthan/api/internal/modules/stock"
)

func registerRoutes(app *fiber.App, ctrl controllers, jwtSecret string) {
	registerHealth(app)
	registerV1(app, ctrl, jwtSecret)
}

func registerHealth(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

func registerV1(app *fiber.App, ctrl controllers, jwtSecret string) {
	api := app.Group("/api/v1")
	authRequired := middleware.Auth(jwtSecret)

	auth.RegisterRoutes(api, ctrl.auth, authRequired)
	address.RegisterRoutes(api, ctrl.address, authRequired)
	contact.RegisterRoutes(api, ctrl.contact, authRequired)
	document.RegisterRoutes(api, ctrl.document, authRequired)
	product.RegisterRoutes(api, ctrl.product, authRequired)
	service.RegisterRoutes(api, ctrl.service, authRequired)
	country.RegisterRoutes(api, ctrl.country, authRequired)
	state.RegisterRoutes(api, ctrl.state, authRequired)
	city.RegisterRoutes(api, ctrl.city, authRequired)
	person.RegisterRoutes(api, ctrl.person, authRequired)
	serviceOrder.RegisterRoutes(api, ctrl.serviceOrder, authRequired)
	stock.RegisterRoutes(api, ctrl.stock, authRequired)
}
