package domain

import "time"

type Role string

const (
	RoleSeller Role = "seller"
	RoleAdmin  Role = "admin"
)

type VehicleStatus string

const (
	VehiclePending   VehicleStatus = "pending"
	VehiclePublished VehicleStatus = "published"
	VehicleSold      VehicleStatus = "sold"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Role         Role      `json:"role"`
	PasswordSalt string    `json:"password_salt"`
	PasswordHash string    `json:"password_hash"`
	IsBlocked    bool      `json:"is_blocked"`
	Warnings     []string  `json:"warnings"`
	CreatedAt    time.Time `json:"created_at"`
}

type TechnicalSpecs struct {
	EngineDisplacement string `json:"engine_displacement"`
	FuelType           string `json:"fuel_type"`
	Horsepower         int    `json:"horsepower"`
	TorqueNM           int    `json:"torque_nm"`
	Transmission       string `json:"transmission"`
	BodyType           string `json:"body_type"`
	Doors              int    `json:"doors"`
	MileageKM          int    `json:"mileage_km"`
	Color              string `json:"color"`
	Traction           string `json:"traction"`
}

type Vehicle struct {
	ID          string         `json:"id"`
	Make        string         `json:"make"`
	Model       string         `json:"model"`
	Year        int            `json:"year"`
	PriceCOP    int64          `json:"price_cop"`
	Description string         `json:"description"`
	Images      []string       `json:"images"`
	Specs       TechnicalSpecs `json:"specs"`
	SellerID    string         `json:"seller_id"`
	Status      VehicleStatus  `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
}

type SellerContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type PublicVehicle struct {
	ID          string         `json:"id"`
	Make        string         `json:"make"`
	Model       string         `json:"model"`
	Year        int            `json:"year"`
	PriceCOP    int64          `json:"price_cop"`
	Description string         `json:"description"`
	Images      []string       `json:"images"`
	Specs       TechnicalSpecs `json:"specs"`
	Status      VehicleStatus  `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	Seller      SellerContact  `json:"seller"`
}
