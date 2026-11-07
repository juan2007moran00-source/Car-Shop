package service

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
	"github.com/juan2007moran00-source/Car-Shop/internal/store"
)

var (
	ErrInvalidInput       = errors.New("datos inválidos")
	ErrInvalidCredentials = errors.New("correo o contraseña inválidos")
	ErrUserBlocked        = errors.New("tu cuenta ha sido bloqueada por un administrador")
)

const passwordIterations = 600_000

type Service struct {
	store     *store.Store
	sessions  map[string]string
	sessionMu sync.RWMutex
}

type RegisterSellerInput struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

type VehicleInput struct {
	Make        string
	Model       string
	Year        int
	PriceCOP    int64
	Description string
	Images      []string
	Specs       domain.TechnicalSpecs
}

func New(dataStore *store.Store) *Service {
	return &Service{store: dataStore, sessions: make(map[string]string)}
}

func (s *Service) EnsureAdmin(email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return errors.New("ADMIN_EMAIL y ADMIN_PASSWORD deben configurarse juntos")
	}
	if !validEmail(email) || len(password) < 8 {
		return errors.New("las credenciales iniciales del administrador no son válidas")
	}
	if existing, ok := s.store.UserByEmail(email); ok {
		if existing.Role != domain.RoleAdmin {
			return errors.New("ADMIN_EMAIL ya pertenece a una cuenta de vendedor")
		}
		return nil
	}
	user, err := newUser("Administrador", email, "", domain.RoleAdmin, password)
	if err != nil {
		return err
	}
	return s.store.CreateUser(user)
}

func (s *Service) RegisterSeller(input RegisterSellerInput) (domain.User, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	phone := strings.TrimSpace(input.Phone)
	if name == "" || len(name) > 120 || !validEmail(email) || phone == "" || len(phone) > 40 || len(input.Password) < 8 {
		return domain.User{}, ErrInvalidInput
	}
	user, err := newUser(name, email, phone, domain.RoleSeller, input.Password)
	if err != nil {
		return domain.User{}, err
	}
	if err := s.store.CreateUser(user); err != nil {
		if errors.Is(err, store.ErrDuplicateUser) {
			return domain.User{}, err
		}
		return domain.User{}, fmt.Errorf("guardar vendedor: %w", err)
	}
	return user, nil
}

func (s *Service) Login(email, password string) (domain.User, string, error) {
	user, ok := s.store.UserByEmail(strings.ToLower(strings.TrimSpace(email)))
	if !ok {
		return domain.User{}, "", ErrInvalidCredentials
	}
	if user.IsBlocked {
		return domain.User{}, "", ErrUserBlocked
	}
	salt, err := hex.DecodeString(user.PasswordSalt)
	if err != nil {
		return domain.User{}, "", ErrInvalidCredentials
	}
	expected, err := hex.DecodeString(user.PasswordHash)
	if err != nil {
		return domain.User{}, "", ErrInvalidCredentials
	}
	actual, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, len(expected))
	if err != nil || subtle.ConstantTimeCompare(actual, expected) != 1 {
		return domain.User{}, "", ErrInvalidCredentials
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return domain.User{}, "", fmt.Errorf("crear sesión: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	s.sessionMu.Lock()
	s.sessions[hex.EncodeToString(tokenHash[:])] = user.ID
	s.sessionMu.Unlock()
	return user, token, nil
}

func (s *Service) Authenticate(token string) (domain.User, bool) {
	tokenHash := sha256.Sum256([]byte(token))
	s.sessionMu.RLock()
	userID, ok := s.sessions[hex.EncodeToString(tokenHash[:])]
	s.sessionMu.RUnlock()
	if !ok {
		return domain.User{}, false
	}
	user, ok := s.store.UserByID(userID)
	if !ok || user.IsBlocked {
		return domain.User{}, false
	}
	return user, true
}

func (s *Service) PublicVehicles() []domain.PublicVehicle {
	vehicles := s.store.VehiclesByStatus(domain.VehiclePublished)
	result := make([]domain.PublicVehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		if publicVehicle, ok := s.withSeller(vehicle); ok {
			result = append(result, publicVehicle)
		}
	}
	return result
}

func (s *Service) PublicVehicle(id string) (domain.PublicVehicle, error) {
	vehicle, ok := s.store.VehicleByID(id)
	if !ok || vehicle.Status != domain.VehiclePublished {
		return domain.PublicVehicle{}, store.ErrNotFound
	}
	publicVehicle, ok := s.withSeller(vehicle)
	if !ok {
		return domain.PublicVehicle{}, store.ErrNotFound
	}
	return publicVehicle, nil
}

func (s *Service) PendingVehicles() []domain.PublicVehicle {
	vehicles := s.store.VehiclesByStatus(domain.VehiclePending)
	result := make([]domain.PublicVehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		if publicVehicle, ok := s.withSeller(vehicle); ok {
			result = append(result, publicVehicle)
		}
	}
	return result
}

func (s *Service) Users() []domain.User {
	return s.store.Users()
}

func validateSpecs(specs domain.TechnicalSpecs) bool {
	if strings.TrimSpace(specs.EngineDisplacement) == "" || len(specs.EngineDisplacement) > 40 {
		return false
	}
	if strings.TrimSpace(specs.FuelType) == "" || len(specs.FuelType) > 40 {
		return false
	}
	if specs.Horsepower <= 0 || specs.Horsepower > 3000 {
		return false
	}
	if specs.TorqueNM <= 0 || specs.TorqueNM > 5000 {
		return false
	}
	if strings.TrimSpace(specs.Transmission) == "" || len(specs.Transmission) > 40 {
		return false
	}
	if strings.TrimSpace(specs.BodyType) == "" || len(specs.BodyType) > 40 {
		return false
	}
	if specs.Doors <= 0 || specs.Doors > 10 {
		return false
	}
	if specs.MileageKM < 0 || specs.MileageKM > 2000000 {
		return false
	}
	if strings.TrimSpace(specs.Color) == "" || len(specs.Color) > 40 {
		return false
	}
	if strings.TrimSpace(specs.Traction) == "" || len(specs.Traction) > 40 {
		return false
	}
	return true
}

func (s *Service) PublishVehicle(sellerID string, input VehicleInput) (domain.Vehicle, error) {
	makeName := strings.TrimSpace(input.Make)
	model := strings.TrimSpace(input.Model)
	description := strings.TrimSpace(input.Description)
	if makeName == "" || model == "" || len(makeName) > 80 || len(model) > 120 ||
		input.Year < 1886 || input.Year > time.Now().Year()+1 || input.PriceCOP <= 0 || len(description) > 1000 {
		return domain.Vehicle{}, ErrInvalidInput
	}
	if len(input.Images) == 0 {
		return domain.Vehicle{}, ErrInvalidInput
	}
	for _, img := range input.Images {
		if strings.TrimSpace(img) == "" {
			return domain.Vehicle{}, ErrInvalidInput
		}
	}
	if !validateSpecs(input.Specs) {
		return domain.Vehicle{}, ErrInvalidInput
	}
	if seller, ok := s.store.UserByID(sellerID); !ok || seller.Role != domain.RoleSeller {
		return domain.Vehicle{}, store.ErrNotFound
	}
	id, err := newID()
	if err != nil {
		return domain.Vehicle{}, err
	}
	vehicle := domain.Vehicle{
		ID:          id,
		Make:        makeName,
		Model:       model,
		Year:        input.Year,
		PriceCOP:    input.PriceCOP,
		Description: description,
		Images:      input.Images,
		Specs:       input.Specs,
		SellerID:    sellerID,
		Status:      domain.VehiclePending,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.store.CreateVehicle(vehicle); err != nil {
		return domain.Vehicle{}, fmt.Errorf("guardar vehículo: %w", err)
	}
	return vehicle, nil
}

func (s *Service) ReportSale(vehicleID, sellerID string) error {
	return s.store.MarkVehicleSold(vehicleID, sellerID)
}

func (s *Service) ApproveVehicle(vehicleID string) error {
	return s.store.ApproveVehicle(vehicleID)
}

func (s *Service) DeleteVehicle(vehicleID string) error {
	return s.store.DeleteVehicle(vehicleID)
}

func (s *Service) DeleteUser(userID string) error {
	if err := s.store.DeleteUser(userID); err != nil {
		return err
	}
	s.sessionMu.Lock()
	for tokenHash, id := range s.sessions {
		if id == userID {
			delete(s.sessions, tokenHash)
		}
	}
	s.sessionMu.Unlock()
	return nil
}

func (s *Service) SellerVehicles(sellerID string) []domain.Vehicle {
	return s.store.VehiclesBySeller(sellerID)
}

func (s *Service) UpdateVehicle(sellerID, vehicleID string, priceCOP int64, description string, images []string, specs domain.TechnicalSpecs) (domain.Vehicle, error) {
	description = strings.TrimSpace(description)
	if priceCOP <= 0 || len(description) > 1000 {
		return domain.Vehicle{}, ErrInvalidInput
	}
	if len(images) > 0 {
		for _, img := range images {
			if strings.TrimSpace(img) == "" {
				return domain.Vehicle{}, ErrInvalidInput
			}
		}
	}
	if specs.EngineDisplacement != "" && !validateSpecs(specs) {
		return domain.Vehicle{}, ErrInvalidInput
	}
	return s.store.UpdateVehicleDetails(vehicleID, sellerID, priceCOP, description, images, specs)
}

func (s *Service) DeleteSellerVehicle(sellerID, vehicleID string) error {
	return s.store.DeleteSellerVehicle(vehicleID, sellerID)
}

func (s *Service) AdminAllVehicles() []domain.PublicVehicle {
	vehicles := s.store.AllVehicles()
	result := make([]domain.PublicVehicle, 0, len(vehicles))
	for _, vehicle := range vehicles {
		if publicVehicle, ok := s.withSeller(vehicle); ok {
			result = append(result, publicVehicle)
		}
	}
	return result
}

func (s *Service) BlockUser(userID string) (domain.User, error) {
	user, err := s.store.SetUserBlocked(userID, true)
	if err != nil {
		return domain.User{}, err
	}
	s.sessionMu.Lock()
	for tokenHash, id := range s.sessions {
		if id == userID {
			delete(s.sessions, tokenHash)
		}
	}
	s.sessionMu.Unlock()
	return user, nil
}

func (s *Service) UnblockUser(userID string) (domain.User, error) {
	return s.store.SetUserBlocked(userID, false)
}

func (s *Service) WarnUser(userID, message string) (domain.User, error) {
	message = strings.TrimSpace(message)
	if message == "" || len(message) > 500 {
		return domain.User{}, ErrInvalidInput
	}
	return s.store.AddUserWarning(userID, message)
}

func (s *Service) withSeller(vehicle domain.Vehicle) (domain.PublicVehicle, bool) {
	seller, ok := s.store.UserByID(vehicle.SellerID)
	if !ok {
		return domain.PublicVehicle{}, false
	}
	images := vehicle.Images
	if images == nil {
		images = []string{}
	}
	return domain.PublicVehicle{
		ID:          vehicle.ID,
		Make:        vehicle.Make,
		Model:       vehicle.Model,
		Year:        vehicle.Year,
		PriceCOP:    vehicle.PriceCOP,
		Description: vehicle.Description,
		Images:      images,
		Specs:       vehicle.Specs,
		Status:      vehicle.Status,
		CreatedAt:   vehicle.CreatedAt,
		Seller: domain.SellerContact{
			Name:  seller.Name,
			Email: seller.Email,
			Phone: seller.Phone,
		},
	}, true
}

func newUser(name, email, phone string, role domain.Role, password string) (domain.User, error) {
	id, err := newID()
	if err != nil {
		return domain.User{}, err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return domain.User{}, err
	}
	passwordHash, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		ID:           id,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Role:         role,
		PasswordSalt: hex.EncodeToString(salt),
		PasswordHash: hex.EncodeToString(passwordHash),
		IsBlocked:    false,
		Warnings:     []string{},
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

func newID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
