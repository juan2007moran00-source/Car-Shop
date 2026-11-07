package store

import (
	"errors"

	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
)

var (
	ErrNotFound      = errors.New("registro no encontrado")
	ErrDuplicateUser = errors.New("el correo ya está registrado")
	ErrInvalidState  = errors.New("el estado actual no permite esta operación")
)

// Repository define la persistencia que consumen los servicios.
// La implementación es PostgreSQL (ver postgres.go).
type Repository interface {
	CreateUser(user domain.User) error
	UserByEmail(email string) (domain.User, bool)
	UserByID(id string) (domain.User, bool)
	Users() []domain.User

	CreateVehicle(vehicle domain.Vehicle) error
	VehicleByID(id string) (domain.Vehicle, bool)
	VehiclesByStatus(status domain.VehicleStatus) []domain.Vehicle
	VehiclesBySeller(sellerID string) []domain.Vehicle
	AllVehicles() []domain.Vehicle
	ApproveVehicle(id string) error
	MarkVehicleSold(id, sellerID string) error
	DeleteVehicle(id string) error
	DeleteUser(id string) error
}
