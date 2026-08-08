package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

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
	"github.com/natthan/api/internal/standalone"
)

//go:embed dist
var distFS embed.FS

const port = "8080"

func main() {
	// ─── Data directory (next to the executable) ──────────────────────────────
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := filepath.Join(filepath.Dir(exePath), "natthan-data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("criando diretório de dados: %v", err)
	}

	// ─── Database ─────────────────────────────────────────────────────────────
	db, err := standalone.OpenDB(dataDir)
	if err != nil {
		log.Fatalf("banco de dados: %v", err)
	}
	defer db.Close()

	standalone.SeedDemoUsers(db)

	// ─── Cache (in-memory, substitui Redis) ───────────────────────────────────
	memCache := standalone.NewMemCache()

	// ─── JWT (gerado aleatoriamente em cada execução) ─────────────────────────
	jwtSecret := randomSecret()
	accessTTL := 15 * time.Minute
	refreshTTL := 168 * time.Hour

	// ─── Wire ─────────────────────────────────────────────────────────────────
	countryRepo := standalone.NewCountryRepository(db)
	countrySvc := country.NewService(countryRepo)
	countryCtrl := country.NewController(countrySvc)

	stateRepo := standalone.NewStateRepository(db)
	stateSvc := state.NewService(stateRepo, countrySvc)
	stateCtrl := state.NewController(stateSvc)

	cityRepo := standalone.NewCityRepository(db)
	citySvc := city.NewService(cityRepo, stateSvc)
	cityCtrl := city.NewController(citySvc)

	addressRepo := standalone.NewAddressRepository(db)
	addressSvc := address.NewService(addressRepo, citySvc)
	addressCtrl := address.NewController(addressSvc)

	contactRepo := standalone.NewContactRepository(db)
	contactSvc := contact.NewService(contactRepo)
	contactCtrl := contact.NewController(contactSvc)

	documentRepo := standalone.NewDocumentRepository(db)
	documentSvc := document.NewService(documentRepo)
	documentCtrl := document.NewController(documentSvc)

	personRepo := standalone.NewPersonRepository(db)
	personSvc := person.NewService(personRepo, addressSvc, contactSvc, documentSvc)
	personCtrl := person.NewController(personSvc)

	authRepo := standalone.NewAuthRepository(db)
	authSvc := auth.NewService(authRepo, personSvc, memCache, jwtSecret, accessTTL, refreshTTL)
	authCtrl := auth.NewController(authSvc)

	productRepo := standalone.NewProductRepository(db)
	productSvc := product.NewService(productRepo)
	productCtrl := product.NewController(productSvc)

	serviceRepo := standalone.NewServiceRepository(db)
	serviceSvc := service.NewService(serviceRepo)
	serviceCtrl := service.NewController(serviceSvc)

	soRepo := standalone.NewServiceOrderRepository(db)
	soSvc := serviceOrder.NewService(soRepo)
	soCtrl := serviceOrder.NewController(soSvc)

	stockRepo := standalone.NewStockRepository(db)
	stockSvc := stock.NewService(stockRepo)
	stockCtrl := stock.NewController(stockSvc)

	// ─── Fiber ────────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Natthan",
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.NewCORS())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")
	authRequired := middleware.Auth(jwtSecret)
	auth.RegisterRoutes(api, authCtrl, authRequired)
	address.RegisterRoutes(api, addressCtrl, authRequired)
	contact.RegisterRoutes(api, contactCtrl, authRequired)
	document.RegisterRoutes(api, documentCtrl, authRequired)
	product.RegisterRoutes(api, productCtrl, authRequired)
	service.RegisterRoutes(api, serviceCtrl, authRequired)
	country.RegisterRoutes(api, countryCtrl, authRequired)
	state.RegisterRoutes(api, stateCtrl, authRequired)
	city.RegisterRoutes(api, cityCtrl, authRequired)
	person.RegisterRoutes(api, personCtrl, authRequired)
	serviceOrder.RegisterRoutes(api, soCtrl, authRequired)
	stock.RegisterRoutes(api, stockCtrl, authRequired)

	// SPA fallback: serve embedded frontend for all non-API routes
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("frontend embed: %v", err)
	}
	app.Use(filesystem.New(filesystem.Config{
		Root:         http.FS(sub),
		Index:        "index.html",
		NotFoundFile: "index.html",
		Browse:       false,
	}))

	// ─── Start ────────────────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Natthan rodando em http://localhost:%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("servidor: %v", err)
		}
	}()

	// Wait briefly then open the browser.
	time.Sleep(500 * time.Millisecond)
	openBrowser("http://localhost:" + port)

	<-quit
	log.Println("encerrando...")
	_ = app.Shutdown()
}

func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
