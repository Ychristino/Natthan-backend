package main

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natthan/api/internal/core/cache"
	"github.com/natthan/api/internal/core/config"
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

type controllers struct {
	auth         *auth.Controller
	product      *product.Controller
	service      *service.Controller
	address      *address.Controller
	contact      *contact.Controller
	document     *document.Controller
	country      *country.Controller
	state        *state.Controller
	city         *city.Controller
	person       *person.Controller
	serviceOrder *serviceOrder.Controller
	stock        *stock.Controller
}

func wire(db *pgxpool.Pool, cacheClient cache.Store, cfg *config.Config) controllers {
	accessTTL, _ := time.ParseDuration(cfg.JWTAccessTokenTTL)
	refreshTTL, _ := time.ParseDuration(cfg.JWTRefreshTokenTTL)

	// ── localization ──────────────────────────────────────────────────────────
	countryRepo := country.NewRepository(db)
	countrySvc := country.NewService(countryRepo)
	countryCtrl := country.NewController(countrySvc)

	stateRepo := state.NewRepository(db)
	stateSvc := state.NewService(stateRepo, countrySvc)
	stateCtrl := state.NewController(stateSvc)

	cityRepo := city.NewRepository(db)
	citySvc := city.NewService(cityRepo, stateSvc)
	cityCtrl := city.NewController(citySvc)

	// ── address ───────────────────────────────────────────────────────────────
	addressRepo := address.NewRepository(db)
	addressSvc := address.NewService(addressRepo, citySvc)
	addressCtrl := address.NewController(addressSvc)

	// ── contact ───────────────────────────────────────────────────────────────
	contactRepo := contact.NewRepository(db)
	contactSvc := contact.NewService(contactRepo)
	contactCtrl := contact.NewController(contactSvc)

	// ── document ──────────────────────────────────────────────────────────────
	documentRepo := document.NewRepository(db)
	documentSvc := document.NewService(documentRepo)
	documentCtrl := document.NewController(documentSvc)

	// ── person ────────────────────────────────────────────────────────────────
	personRepo := person.NewRepository(db)
	personSvc := person.NewService(personRepo, addressSvc, contactSvc, documentSvc)
	personCtrl := person.NewController(personSvc)

	// ── auth ──────────────────────────────────────────────────────────────────
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, personSvc, cacheClient, cfg.JWTSecret, accessTTL, refreshTTL)
	authCtrl := auth.NewController(authSvc)

	// ── product ───────────────────────────────────────────────────────────────
	productRepo := product.NewRepository(db)
	productSvc := product.NewService(productRepo)
	productCtrl := product.NewController(productSvc)

	// ── service ───────────────────────────────────────────────────────────────
	serviceRepo := service.NewRepository(db)
	serviceSvc := service.NewService(serviceRepo)
	serviceCtrl := service.NewController(serviceSvc)

	// ── serviceOrder ──────────────────────────────────────────────────────────
	serviceOrderRepo := serviceOrder.NewRepository(db)
	serviceOrderSvc := serviceOrder.NewService(serviceOrderRepo)
	serviceOrderCtrl := serviceOrder.NewController(serviceOrderSvc)

	// ── stock ─────────────────────────────────────────────────────────────────
	stockRepo := stock.NewRepository(db)
	stockSvc := stock.NewService(stockRepo)
	stockCtrl := stock.NewController(stockSvc)

	return controllers{
		auth:         authCtrl,
		product:      productCtrl,
		service:      serviceCtrl,
		address:      addressCtrl,
		contact:      contactCtrl,
		document:     documentCtrl,
		country:      countryCtrl,
		state:        stateCtrl,
		city:         cityCtrl,
		person:       personCtrl,
		serviceOrder: serviceOrderCtrl,
		stock:        stockCtrl,
	}
}
