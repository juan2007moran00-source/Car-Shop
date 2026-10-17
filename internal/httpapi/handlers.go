package httpapi

import (
	"errors"
	"net/http"
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
	Make        string `json:"make"`
	Model       string `json:"model"`
	Year        int    `json:"year"`
	PriceCOP    int64  `json:"price_cop"`
	Description string `json:"description"`
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
		c.JSON(http.StatusOK, gin.H{"token": token, "token_type": "Bearer", "user": gin.H{
			"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role,
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
			Make: request.Make, Model: request.Model, Year: request.Year,
			PriceCOP: request.PriceCOP, Description: request.Description,
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
	case errors.Is(err, store.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso no encontrado"})
	case errors.Is(err, store.ErrInvalidState):
		c.JSON(http.StatusConflict, gin.H{"error": "el estado actual no permite esta operación"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar la operación"})
	}
}
