package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juan2007moran00-source/Car-Shop/internal/media"
)

func serveMedia(mediaStore *media.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		path, ok := mediaStore.Resolve(c.Param("vehicleID"), c.Param("fileName"))
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "public, max-age=604800, immutable")
		c.File(path)
	}
}
