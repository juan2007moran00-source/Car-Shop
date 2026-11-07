package httpapi

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
	"github.com/juan2007moran00-source/Car-Shop/internal/media"
	"github.com/juan2007moran00-source/Car-Shop/internal/service"
	"github.com/juan2007moran00-source/Car-Shop/internal/store"
)

type registerSellerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func registerSeller(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request registerSellerRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el cuerpo debe ser JSON válido"})
			return
		}
		user, err := app.RegisterSeller(service.RegisterSellerInput{
			Name: request.Name, Email: request.Email, Phone: request.Phone, Password: request.Password,
		})
		if err != nil {
			writeServiceError(c, err)
			return
		}
		setActor(c, user)
		c.JSON(http.StatusCreated, gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role})
	}
}

func login(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request loginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el cuerpo debe ser JSON válido"})
			return
		}
		user, token, err := app.Login(request.Email, request.Password)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		setActor(c, user)
		c.JSON(http.StatusOK, gin.H{"token": token, "token_type": "Bearer", "user": userPayload(user)})
	}
}

func currentSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": userPayload(currentUser(c))})
	}
}

func logout(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, ok := c.Get("session_token"); ok {
			if value, ok := token.(string); ok {
				app.Logout(value)
			}
		}
		c.Status(http.StatusNoContent)
	}
}

func listPublicVehicles(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"vehicles": app.PublicVehicles()})
	}
}

func getPublicVehicle(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		vehicle, err := app.PublicVehicle(c.Param("id"))
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, vehicle)
	}
}

func publishVehicle(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)

		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo procesar el formulario"})
			return
		}
		year, err := strconv.Atoi(strings.TrimSpace(c.PostForm("year")))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el año no es válido"})
			return
		}
		price, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("price_cop")), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el precio no es válido"})
			return
		}
		coverIndex := 0
		if rawCover := strings.TrimSpace(c.PostForm("cover_index")); rawCover != "" {
			coverIndex, err = strconv.Atoi(rawCover)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "la portada seleccionada no es válida"})
				return
			}
		}
		input := service.VehicleInput{
			Make:        c.PostForm("make"),
			Model:       c.PostForm("model"),
			Year:        year,
			PriceCOP:    price,
			Description: c.PostForm("description"),
		}

		form := c.Request.MultipartForm
		files := form.File["images"]
		if len(files) == 0 {
			files = form.File["images[]"]
		}
		var opened []multipart.File
		defer func() {
			for _, file := range opened {
				file.Close()
			}
		}()
		uploads := make([]service.ImageUpload, 0, len(files))
		for _, header := range files {
			file, err := header.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo leer una de las imágenes"})
				return
			}
			opened = append(opened, file)
			uploads = append(uploads, service.ImageUpload{Name: header.Filename, Reader: file})
		}

		vehicle, err := app.PublishVehicle(user.ID, input, uploads, coverIndex)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusCreated, vehicle)
	}
}

func listSellerVehicles(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		c.JSON(http.StatusOK, gin.H{"vehicles": app.SellerVehicles(user.ID)})
	}
}

func reportSale(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if err := app.ReportSale(c.Param("id"), user.ID); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func listPendingVehicles(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"vehicles": app.PendingVehicles()})
	}
}

func listAllVehicles(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"vehicles": app.AllVehicles()})
	}
}

func listUsers(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		users := app.Users()
		result := make([]gin.H, 0, len(users))
		for _, user := range users {
			result = append(result, gin.H{
				"id": user.ID, "name": user.Name, "email": user.Email, "phone": user.Phone,
				"role": user.Role, "created_at": user.CreatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"users": result})
	}
}

func approveVehicle(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := app.ApproveVehicle(c.Param("id")); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteVehicle(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := app.DeleteVehicle(c.Param("id")); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteUser(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if currentUser(c).ID == c.Param("id") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no puedes eliminar tu propia cuenta"})
			return
		}
		if err := app.DeleteUser(c.Param("id")); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func authentication(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "se requiere autenticación"})
			return
		}
		user, ok := app.Authenticate(parts[1])
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesión inválida"})
			return
		}
		c.Set("session_token", parts[1])
		setActor(c, user)
		c.Set("current_user", user)
		c.Next()
	}
}

func requireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if string(user.Role) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no tienes permiso para esta operación"})
			return
		}
		c.Next()
	}
}

func currentUser(c *gin.Context) domain.User {
	value, _ := c.Get("current_user")
	user, _ := value.(domain.User)
	return user
}

func setActor(c *gin.Context, user domain.User) {
	c.Set("actor_id", user.ID)
	c.Set("actor_role", string(user.Role))
}

func userPayload(user domain.User) gin.H {
	return gin.H{
		"id": user.ID, "name": user.Name, "email": user.Email,
		"phone": user.Phone, "role": user.Role, "created_at": user.CreatedAt,
	}
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, media.ErrUnsupportedImage),
		errors.Is(err, media.ErrInvalidImage):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "los datos enviados no son válidos"})
	case errors.Is(err, store.ErrDuplicateUser):
		c.JSON(http.StatusConflict, gin.H{"error": "ya existe una cuenta con ese correo"})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "correo o contraseña inválidos"})
	case errors.Is(err, store.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso no encontrado"})
	case errors.Is(err, store.ErrInvalidState):
		c.JSON(http.StatusConflict, gin.H{"error": "el estado actual no permite esta operación"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar la operación"})
	}
}
