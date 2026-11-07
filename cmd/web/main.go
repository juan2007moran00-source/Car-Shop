package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/juan2007moran00-source/Car-Shop/internal/audit"
	"github.com/juan2007moran00-source/Car-Shop/internal/httpapi"
	"github.com/juan2007moran00-source/Car-Shop/internal/media"
	"github.com/juan2007moran00-source/Car-Shop/internal/service"
	"github.com/juan2007moran00-source/Car-Shop/internal/store"
)

func main() {
	databaseURL := strings.TrimSpace(os.Getenv("AUTOMARKET_DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("AUTOMARKET_DATABASE_URL es obligatoria: define la cadena de conexión a PostgreSQL")
	}
	auditPath := envOr("AUTOMARKET_AUDIT_FILE", "data/audit.jsonl")
	mediaPath := envOr("AUTOMARKET_MEDIA_DIR", "data/uploads")
	certDir := envOr("AUTOMARKET_CERT_DIR", "data/certs")
	address := envOr("AUTOMARKET_ADDR", ":8443")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	dataStore, err := store.NewPostgres(ctx, databaseURL)
	cancel()
	if err != nil {
		log.Fatalf("no se pudo conectar con PostgreSQL: %v", err)
	}
	defer dataStore.Close()

	auditLogger, err := audit.NewLogger(auditPath)
	if err != nil {
		log.Fatalf("no se pudo cargar la auditoría: %v", err)
	}
	mediaStore, err := media.New(mediaPath)
	if err != nil {
		log.Fatalf("no se pudo preparar el almacenamiento de imágenes: %v", err)
	}
	app := service.New(dataStore, mediaStore)
	if err := app.EnsureAdmin(os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")); err != nil {
		log.Fatalf("no se pudo configurar el administrador: %v", err)
	}

	certFile, keyFile, err := ensureDevCertificate(certDir)
	if err != nil {
		log.Fatalf("no se pudo preparar el certificado HTTPS: %v", err)
	}

	router := httpapi.NewRouter(app, auditLogger, mediaStore)
	log.Printf("AutoMarket escuchando en https://localhost%s", address)
	if err := router.RunTLS(address, certFile, keyFile); err != nil {
		log.Fatalf("el servidor HTTPS no pudo iniciar: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
