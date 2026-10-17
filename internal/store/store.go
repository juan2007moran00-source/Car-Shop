package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
)

var (
	ErrNotFound      = errors.New("registro no encontrado")
	ErrDuplicateUser = errors.New("el correo ya está registrado")
	ErrInvalidState  = errors.New("el estado actual no permite esta operación")
)

type Store struct {
	mu       sync.RWMutex
	path     string
	users    map[string]domain.User
	vehicles map[string]domain.Vehicle
}

type snapshot struct {
	Users    []domain.User    `json:"users"`
	Vehicles []domain.Vehicle `json:"vehicles"`
}

func New(path string) (*Store, error) {
	s := &Store{
		path:     path,
		users:    make(map[string]domain.User),
		vehicles: make(map[string]domain.Vehicle),
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}

	var saved snapshot
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}
	for _, user := range saved.Users {
		s.users[user.ID] = user
	}
	for _, vehicle := range saved.Vehicles {
		s.vehicles[vehicle.ID] = vehicle
	}
	return s, nil
}

func (s *Store) CreateUser(user domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if strings.EqualFold(existing.Email, user.Email) {
			return ErrDuplicateUser
		}
	}
	s.users[user.ID] = user
	if err := s.saveLocked(); err != nil {
		delete(s.users, user.ID)
		return err
	}
	return nil
}

func (s *Store) UserByEmail(email string) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if strings.EqualFold(user.Email, email) {
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *Store) UserByID(id string) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

func (s *Store) Users() []domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].Email < users[j].Email })
	return users
}

func (s *Store) CreateVehicle(vehicle domain.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vehicles[vehicle.ID] = vehicle
	if err := s.saveLocked(); err != nil {
		delete(s.vehicles, vehicle.ID)
		return err
	}
	return nil
}

func (s *Store) VehicleByID(id string) (domain.Vehicle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicle, ok := s.vehicles[id]
	return vehicle, ok
}

func (s *Store) VehiclesByStatus(status domain.VehicleStatus) []domain.Vehicle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicles := make([]domain.Vehicle, 0)
	for _, vehicle := range s.vehicles {
		if vehicle.Status == status {
			vehicles = append(vehicles, vehicle)
		}
	}
	sort.Slice(vehicles, func(i, j int) bool {
		return vehicles[i].CreatedAt.After(vehicles[j].CreatedAt)
	})
	return vehicles
}

func (s *Store) ApproveVehicle(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vehicle, ok := s.vehicles[id]
	if !ok {
		return ErrNotFound
	}
	if vehicle.Status != domain.VehiclePending {
		return ErrInvalidState
	}
	vehicle.Status = domain.VehiclePublished
	s.vehicles[id] = vehicle
	if err := s.saveLocked(); err != nil {
		vehicle.Status = domain.VehiclePending
		s.vehicles[id] = vehicle
		return err
	}
	return nil
}

func (s *Store) MarkVehicleSold(id, sellerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vehicle, ok := s.vehicles[id]
	if !ok {
		return ErrNotFound
	}
	if vehicle.SellerID != sellerID {
		return ErrNotFound
	}
	if vehicle.Status != domain.VehiclePublished {
		return ErrInvalidState
	}
	vehicle.Status = domain.VehicleSold
	s.vehicles[id] = vehicle
	if err := s.saveLocked(); err != nil {
		vehicle.Status = domain.VehiclePublished
		s.vehicles[id] = vehicle
		return err
	}
	return nil
}

func (s *Store) DeleteVehicle(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vehicle, ok := s.vehicles[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.vehicles, id)
	if err := s.saveLocked(); err != nil {
		s.vehicles[id] = vehicle
		return err
	}
	return nil
}

func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return ErrNotFound
	}
	removedVehicles := make(map[string]domain.Vehicle)
	for vehicleID, vehicle := range s.vehicles {
		if vehicle.SellerID == id {
			removedVehicles[vehicleID] = vehicle
			delete(s.vehicles, vehicleID)
		}
	}
	delete(s.users, id)
	if err := s.saveLocked(); err != nil {
		s.users[id] = user
		for vehicleID, vehicle := range removedVehicles {
			s.vehicles[vehicleID] = vehicle
		}
		return err
	}
	return nil
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	saved := snapshot{
		Users:    make([]domain.User, 0, len(s.users)),
		Vehicles: make([]domain.Vehicle, 0, len(s.vehicles)),
	}
	for _, user := range s.users {
		saved.Users = append(saved.Users, user)
	}
	for _, vehicle := range s.vehicles {
		saved.Vehicles = append(saved.Vehicles, vehicle)
	}
	sort.Slice(saved.Users, func(i, j int) bool { return saved.Users[i].ID < saved.Users[j].ID })
	sort.Slice(saved.Vehicles, func(i, j int) bool { return saved.Vehicles[i].ID < saved.Vehicles[j].ID })

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(s.path), ".automarket-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, s.path)
}
