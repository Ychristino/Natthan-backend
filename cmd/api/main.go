package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/natthan/api/internal/core/cache"
	"github.com/natthan/api/internal/core/config"
	"github.com/natthan/api/internal/core/database"
	"github.com/natthan/api/internal/middleware"
	"github.com/natthan/api/internal/standalone"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	ctx := context.Background()

	// ─── Infraestrutura ───────────────────────────────────────────────────────

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("banco: %v", err)
	}
	defer db.Close()

	var cacheClient cache.Store
	if c, err := cache.New(cfg.RedisURL); err != nil {
		log.Printf("redis: %v; falling back to in-memory cache", err)
		cacheClient = standalone.NewMemCache()
	} else {
		cacheClient = c
	}

	_ = cacheClient // será usado pelos services após definição do banco

	ctrl := wire(db, cacheClient, cfg)

	// ─── Fiber ────────────────────────────────────────────────────────────────

	app := fiber.New(fiber.Config{
		AppName:      "Natthan API",
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.NewCORS())

	registerRoutes(app, ctrl, cfg.JWTSecret)

	// ─── Graceful shutdown ────────────────────────────────────────────────────

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("servidor rodando na porta %s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("erro ao iniciar servidor: %v", err)
		}
	}()

	<-quit
	log.Println("desligando servidor...")
	_ = app.Shutdown()
}
