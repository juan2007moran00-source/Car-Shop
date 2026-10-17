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
	CreatedAt    time.Time `json:"created_at"`
}

type Vehicle struct {
	ID          string        `json:"id"`
	Make        string        `json:"make"`
	Model       string        `json:"model"`
	Year        int           `json:"year"`
	PriceCOP    int64         `json:"price_cop"`
	Description string        `json:"description"`
	SellerID    string        `json:"seller_id"`
	Status      VehicleStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
}

type SellerContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type PublicVehicle struct {
	ID          string        `json:"id"`
	Make        string        `json:"make"`
	Model       string        `json:"model"`
	Year        int           `json:"year"`
	PriceCOP    int64         `json:"price_cop"`
	Description string        `json:"description"`
	Status      VehicleStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	Seller      SellerContact `json:"seller"`
}
