package auth

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, ctrl *Controller, authRequired fiber.Handler) {
	a := router.Group("/auth")
	a.Post("/login", ctrl.Login)
	a.Post("/register", ctrl.Register)
	a.Post("/refresh", ctrl.Refresh)
	a.Post("/logout", ctrl.Logout)

	u := router.Group("/users", authRequired)
	u.Get("/", ctrl.ListUsers)
	u.Get("/:id", ctrl.GetUser)
	u.Post("/:id/roles", ctrl.AddRole)
	u.Delete("/:id/roles/:role", ctrl.RemoveRole)
	u.Delete("/:id", ctrl.DeactivateUser)
	u.Patch("/:id/activate", ctrl.ReactivateUser)
}
