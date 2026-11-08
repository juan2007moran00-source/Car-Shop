package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

	port := os.Getenv("PORT")

	// Pool único de PostgreSQL compartido para todo el aplicación
	// Contexto de 30s que se cancela al finalizar el main
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("configurar PostgreSQL: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("conectar con PostgreSQL: %v", err)
	}

	// Configuración común de almacenes usando el mismo pool
	dataStore, err := store.NewPostgresFromPool(ctx, pool)
	if err != nil {
		log.Fatalf("configurar store de PostgreSQL: %v", err)
	}
	auditLogger, err := audit.NewLogger(pool)
	if err != nil {
		log.Fatalf("no se pudo cargar la auditoría: %v", err)
	}
	mediaStore, err := media.New(pool)
	if err != nil {
		log.Fatalf("no se pudo preparar el almacenamiento de imágenes: %v", err)
	}

	app := service.New(dataStore, mediaStore)
	if err := app.EnsureAdmin(os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")); err != nil {
		log.Fatalf("no se pudo configurar el administrador: %v", err)
	}

	if port != "" {
		// Modo Render: HTTP plano en 0.0.0.0:PORT, sin certificado TLS
		gin.SetMode(gin.ReleaseMode)
		log.Printf("AutoMarket escuchando en http://0.0.0.0:%s (modo Render)", port)
		if err := httpapi.NewRouter(app, auditLogger, mediaStore).Run("0.0.0.0:" + port); err != nil {
			log.Fatalf("el servidor HTTP no pudo iniciar: %v", err)
		}
		return
	}

	// Comportamiento local actual: HTTPS con certificado autofirmado en :8443
	localAddress := ":8443"
	if v := os.Getenv("AUTOMARKET_ADDR"); v != "" {
		localAddress = v
	}
	certDir := "data/certs"
	if v := os.Getenv("AUTOMARKET_CERT_DIR"); v != "" {
		certDir = v
	}
	certFile, keyFile, err := ensureDevCertificate(certDir)
	if err != nil {
		log.Fatalf("no se pudo preparar el certificado HTTPS: %v", err)
	}
	log.Printf("AutoMarket escuchando en https://localhost%s", localAddress)
	if err := httpapi.NewRouter(app, auditLogger, mediaStore).RunTLS(localAddress, certFile, keyFile); err != nil {
		log.Fatalf("el servidor HTTPS no pudo iniciar: %v", err)
	}
}
