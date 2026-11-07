package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
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

type publishVehicleRequest struct {
	Make        string                `json:"make"`
	Model       string                `json:"model"`
	Year        int                   `json:"year"`
	PriceCOP    int64                 `json:"price_cop"`
	Description string                `json:"description"`
	Images      []string              `json:"images"`
	Specs       domain.TechnicalSpecs `json:"specs"`
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
		warnings := user.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "token_type": "Bearer", "user": gin.H{
			"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role,
			"is_blocked": user.IsBlocked, "warnings": warnings,
		}})
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
		var request publishVehicleRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el cuerpo debe ser JSON válido"})
			return
		}
		user := currentUser(c)
		vehicle, err := app.PublishVehicle(user.ID, service.VehicleInput{
			Make:        request.Make,
			Model:       request.Model,
			Year:        request.Year,
			PriceCOP:    request.PriceCOP,
			Description: request.Description,
			Images:      request.Images,
			Specs:       request.Specs,
		})
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusCreated, vehicle)
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

func listUsers(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		users := app.Users()
		result := make([]gin.H, 0, len(users))
		for _, user := range users {
			warnings := user.Warnings
			if warnings == nil {
				warnings = []string{}
			}
			result = append(result, gin.H{
				"id": user.ID, "name": user.Name, "email": user.Email, "phone": user.Phone,
				"role": user.Role, "is_blocked": user.IsBlocked, "warnings": warnings, "created_at": user.CreatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"users": result})
	}
}

func getMe(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		warnings := user.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"phone":      user.Phone,
			"role":       user.Role,
			"is_blocked": user.IsBlocked,
			"warnings":   warnings,
			"created_at": user.CreatedAt,
		})
	}
}

func listSellerVehicles(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		c.JSON(http.StatusOK, gin.H{"vehicles": app.SellerVehicles(user.ID)})
	}
}

type updateVehicleRequest struct {
	PriceCOP    int64                 `json:"price_cop"`
	Description string                `json:"description"`
	Images      []string              `json:"images"`
	Specs       domain.TechnicalSpecs `json:"specs"`
}

func updateSellerVehicle(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request updateVehicleRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el cuerpo debe ser JSON válido"})
			return
		}
		user := currentUser(c)
		vehicle, err := app.UpdateVehicle(user.ID, c.Param("id"), request.PriceCOP, request.Description, request.Images, request.Specs)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, vehicle)
	}
}

func uploadImages() gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere un formulario multipart con imágenes"})
			return
		}
		files := form.File["images"]
		if len(files) == 0 {
			files = form.File["files"]
		}
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "debes seleccionar al menos una imagen"})
			return
		}

		_ = os.MkdirAll("data/uploads", 0o755)
		urls := make([]string, 0, len(files))

		for _, file := range files {
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "solo se permiten formatos JPG, JPEG, PNG y WEBP"})
				return
			}
			if file.Size > 15<<20 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "el tamaño máximo por imagen es de 15 MB"})
				return
			}

			randomBytes := make([]byte, 16)
			if _, err := rand.Read(randomBytes); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "error generando nombre de archivo"})
				return
			}
			fileName := hex.EncodeToString(randomBytes) + ext
			dst := filepath.Join("data/uploads", fileName)

			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar la imagen"})
				return
			}
			urls = append(urls, "/uploads/"+fileName)
		}

		c.JSON(http.StatusOK, gin.H{"urls": urls})
	}
}

func deleteSellerVehicle(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if err := app.DeleteSellerVehicle(user.ID, c.Param("id")); err != nil {
			writeServiceError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func listAllVehiclesAdmin(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"vehicles": app.AdminAllVehicles()})
	}
}

type warnUserRequest struct {
	Message string `json:"message"`
}

func warnUser(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request warnUserRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "el cuerpo debe ser JSON válido"})
			return
		}
		user, err := app.WarnUser(c.Param("id"), request.Message)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		warnings := user.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"is_blocked": user.IsBlocked,
			"warnings":   warnings,
		})
	}
}

func blockUser(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := app.BlockUser(c.Param("id"))
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"is_blocked": user.IsBlocked,
		})
	}
}

func unblockUser(app *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := app.UnblockUser(c.Param("id"))
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"is_blocked": user.IsBlocked,
		})
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

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "los datos enviados no son válidos"})
	case errors.Is(err, store.ErrDuplicateUser):
		c.JSON(http.StatusConflict, gin.H{"error": "ya existe una cuenta con ese correo"})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "correo o contraseña inválidos"})
	case errors.Is(err, service.ErrUserBlocked):
		c.JSON(http.StatusForbidden, gin.H{"error": "tu cuenta ha sido bloqueada por un administrador"})
	case errors.Is(err, store.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso no encontrado"})
	case errors.Is(err, store.ErrInvalidState):
		c.JSON(http.StatusConflict, gin.H{"error": "el estado actual no permite esta operación"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar la operación"})
	}
}
