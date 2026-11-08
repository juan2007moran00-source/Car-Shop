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

// VehicleImage describe una imagen almacenada para un vehículo.
type VehicleImage struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	SizeBytes   int64  `json:"size_bytes"`
	Data        []byte `json:"-"`
}

type Vehicle struct {
	ID           string         `json:"id"`
	Make         string         `json:"make"`
	Model        string         `json:"model"`
	Year         int            `json:"year"`
	PriceCOP     int64          `json:"price_cop"`
	Description  string         `json:"description"`
	SellerID     string         `json:"seller_id"`
	Status       VehicleStatus  `json:"status"`
	Images       []VehicleImage `json:"images,omitempty"`
	CoverImageID string         `json:"cover_image_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type SellerContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type PublicImage struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	SizeBytes   int64  `json:"size_bytes"`
	Cover       bool   `json:"cover"`
}

type PublicVehicle struct {
	ID           string        `json:"id"`
	Make         string        `json:"make"`
	Model        string        `json:"model"`
	Year         int           `json:"year"`
	PriceCOP     int64         `json:"price_cop"`
	Description  string        `json:"description"`
	Status       VehicleStatus `json:"status"`
	Images       []PublicImage `json:"images"`
	CoverImageID string        `json:"cover_image_id,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	Seller       SellerContact `json:"seller"`
}
