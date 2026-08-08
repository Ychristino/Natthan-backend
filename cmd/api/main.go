package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	"github.com/natthan/api/internal/core/cache"
	"github.com/natthan/api/internal/core/config"
	"github.com/natthan/api/internal/core/database"
	"github.com/natthan/api/internal/middleware"
	"github.com/natthan/api/internal/standalone"
)

func runMigrations(databaseURL string) error {
	candidates := []string{}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "migrations"),
			filepath.Join(exeDir, "db", "migrations"),
		)
	}

	candidates = append(candidates,
		"./migrations",
		"./db/migrations",
		"/migrations",
		"/app/db/migrations",
	)

	var migrationsDir string
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			migrationsDir = candidate
			break
		}
	}

	if migrationsDir == "" {
		cwd, _ := os.Getwd()
		return fmt.Errorf("migrations directory not found; cwd=%s, candidates=%v", cwd, candidates)
	}

	migrationsPath := "file://" + migrationsDir

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

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
