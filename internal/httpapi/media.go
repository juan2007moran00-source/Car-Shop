package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juan2007moran00-source/Car-Shop/internal/media"
)

func serveMedia(mediaStore *media.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		vehicleID := c.Param("vehicleID")
		fileName := c.Param("fileName")

		data, contentType, ok := mediaStore.Resolve(vehicleID, fileName)
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Content-Type", contentType)
		c.Header("Cache-Control", "public, max-age=604800, immutable")
		c.Status(http.StatusOK)
		c.Writer.Write(data)
	}
}
