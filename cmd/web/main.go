package main

import (
	"log"
	"os"

	"github.com/juan2007moran00-source/Car-Shop/internal/audit"
	"github.com/juan2007moran00-source/Car-Shop/internal/httpapi"
	"github.com/juan2007moran00-source/Car-Shop/internal/service"
	"github.com/juan2007moran00-source/Car-Shop/internal/store"
)

func main() {
	dataPath := envOr("AUTOMARKET_DATA_FILE", "data/automarket.json")
	auditPath := envOr("AUTOMARKET_AUDIT_FILE", "data/audit.jsonl")
	address := envOr("AUTOMARKET_ADDR", ":8080")

	dataStore, err := store.New(dataPath)
	if err != nil {
		log.Fatalf("no se pudo cargar la persistencia: %v", err)
	}
	auditLogger, err := audit.NewLogger(auditPath)
	if err != nil {
		log.Fatalf("no se pudo cargar la auditoría: %v", err)
	}
	app := service.New(dataStore)
	adminEmail := envOr("ADMIN_EMAIL", "admin@automarket.local")
	adminPassword := envOr("ADMIN_PASSWORD", "AdminAutoMarket2026*")
	if err := app.EnsureAdmin(adminEmail, adminPassword); err != nil {
		log.Fatalf("no se pudo configurar el administrador: %v", err)
	}

	router := httpapi.NewRouter(app, auditLogger)
	log.Printf("AutoMarket escuchando en http://localhost%s", address)
	if err := router.Run(address); err != nil {
		log.Fatalf("el servidor no pudo iniciar: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
