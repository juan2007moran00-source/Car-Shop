package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juan2007moran00-source/Car-Shop/internal/audit"
	"github.com/juan2007moran00-source/Car-Shop/internal/service"
)

func NewRouter(app *service.Service, auditLogger *audit.Logger) *gin.Engine {
	router := gin.New()
	router.Use(auditMiddleware(auditLogger), gin.Recovery(), requestSizeLimit())
	router.LoadHTMLFiles("web/templates/inicio.html")
	router.Static("/static", "web/static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "inicio.html", gin.H{"title": "AutoMarket"})
	})
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})

	api := router.Group("/api")
	{
		api.POST("/sellers", registerSeller(app))
		api.POST("/sessions", login(app))
		api.GET("/vehicles", listPublicVehicles(app))
		api.GET("/vehicles/:id", getPublicVehicle(app))

		seller := api.Group("")
		seller.Use(authentication(app), requireRole("seller"))
		seller.POST("/vehicles", publishVehicle(app))
		seller.POST("/vehicles/:id/sale", reportSale(app))

		admin := api.Group("/admin")
		admin.Use(authentication(app), requireRole("admin"))
		admin.GET("/users", listUsers(app))
		admin.GET("/vehicles/pending", listPendingVehicles(app))
		admin.POST("/vehicles/:id/approve", approveVehicle(app))
		admin.DELETE("/vehicles/:id", deleteVehicle(app))
		admin.DELETE("/users/:id", deleteUser(app))
	}
	return router
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
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	}
}
