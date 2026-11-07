package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juan2007moran00-source/Car-Shop/internal/audit"
	"github.com/juan2007moran00-source/Car-Shop/internal/media"
	"github.com/juan2007moran00-source/Car-Shop/internal/service"
)

const maxRequestBodyBytes = 40 << 20

func NewRouter(app *service.Service, auditLogger *audit.Logger, mediaStore *media.Store) *gin.Engine {
	router := gin.New()
	router.Use(securityHeaders(), auditMiddleware(auditLogger), gin.Recovery(), requestSizeLimit())
	router.LoadHTMLFiles("web/templates/inicio.html", "web/templates/login.html")
	router.Static("/static", "web/static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "inicio.html", gin.H{"title": "AutoMarket"})
	})
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "Iniciar sesión | AutoMarket"})
	})
	router.GET("/media/:vehicleID/:fileName", serveMedia(mediaStore))
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})

	api := router.Group("/api")
	{
		api.POST("/sellers", registerSeller(app))
		api.POST("/sessions", login(app))
		api.GET("/vehicles", listPublicVehicles(app))
		api.GET("/vehicles/:id", getPublicVehicle(app))

		authenticated := api.Group("")
		authenticated.Use(authentication(app))
		authenticated.GET("/session", currentSession())
		authenticated.DELETE("/session", logout(app))

		seller := api.Group("")
		seller.Use(authentication(app), requireRole("seller"))
		seller.GET("/seller/vehicles", listSellerVehicles(app))
		seller.POST("/vehicles", publishVehicle(app))
		seller.POST("/vehicles/:id/sale", reportSale(app))

		admin := api.Group("/admin")
		admin.Use(authentication(app), requireRole("admin"))
		admin.GET("/users", listUsers(app))
		admin.GET("/vehicles", listAllVehicles(app))
		admin.GET("/vehicles/pending", listPendingVehicles(app))
		admin.POST("/vehicles/:id/approve", approveVehicle(app))
		admin.DELETE("/vehicles/:id", deleteVehicle(app))
		admin.DELETE("/users/:id", deleteUser(app))
	}
	return router
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func auditMiddleware(logger *audit.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		actorID, _ := c.Get("actor_id")
		role, _ := c.Get("actor_role")
		event := audit.Event{
			Time:   time.Now(),
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Status: c.Writer.Status(),
			Role:   "visitor",
		}
		if value, ok := actorID.(string); ok {
			event.ActorID = value
		}
		if value, ok := role.(string); ok {
			event.Role = value
		}
		if err := logger.Record(event); err != nil {
			gin.DefaultErrorWriter.Write([]byte("no se pudo registrar operación de auditoría: " + err.Error() + "\n"))
		}
	}
}

func requestSizeLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		c.Next()
	}
}
