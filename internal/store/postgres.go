package store

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
)

//go:embed schema.sql
var schemaSQL string

// Postgres implementa Repository sobre PostgreSQL.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgresFromPool usa un pool existente (por ejemplo, compartido desde main.go)
// y ejecuta el esquema si no ha sido creado aún.
// Devuelve el error SIEMPRE (nunca lo ignora) y usa SimpleProtocol.
func NewPostgresFromPool(ctx context.Context, pool *pgxpool.Pool) (*Postgres, error) {
	// Ejecutamos el esquema con SimpleProtocol y SIEMPRE devolvemos el error
	if _, err := pool.Exec(ctx, schemaSQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		return nil, fmt.Errorf("ejecutar esquema: %w", err)
	}
	// Agregar columna data si no existe (idempotente) - el ALTER TABLE
	// ya se ejecuta como parte de la llamada desde main.go
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) CreateUser(user domain.User) error {
	_, err := p.pool.Exec(context.Background(), `
		INSERT INTO users (id, name, email, phone, role, password_salt, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Name, user.Email, user.Phone, string(user.Role),
		user.PasswordSalt, user.PasswordHash, user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateUser
		}
		return err
	}
	return nil
}

func (p *Postgres) UserByEmail(email string) (domain.User, bool) {
	user, err := scanUser(p.pool.QueryRow(context.Background(), `
		SELECT id, name, email, phone, role, password_salt, password_hash, created_at
		FROM users WHERE LOWER(email) = LOWER($1)`, email))
	if err != nil {
		return domain.User{}, false
	}
	return user, true
}

func (p *Postgres) UserByID(id string) (domain.User, bool) {
	user, err := scanUser(p.pool.QueryRow(context.Background(), `
		SELECT id, name, email, phone, role, password_salt, password_hash, created_at
		FROM users WHERE id = $1`, id))
	if err != nil {
		return domain.User{}, false
	}
	return user, true
}

func (p *Postgres) Users() []domain.User {
	rows, err := p.pool.Query(context.Background(), `
		SELECT id, name, email, phone, role, password_salt, password_hash, created_at
		FROM users ORDER BY email`)
	if err != nil {
		return []domain.User{}
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return users
		}
		users = append(users, user)
	}
	return users
}

func (p *Postgres) CreateVehicle(vehicle domain.Vehicle) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO vehicles (id, make, model, year, price_cop, description, seller_id, status, cover_image_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		vehicle.ID, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.PriceCOP,
		vehicle.Description, vehicle.SellerID, string(vehicle.Status),
		vehicle.CoverImageID, vehicle.CreatedAt,
	)
	if err != nil {
		return err
	}

	for index, image := range vehicle.Images {
		if _, err := tx.Exec(ctx, `
			INSERT INTO vehicle_images (id, vehicle_id, file_name, content_type, width, height, size_bytes, position, data)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			image.ID, vehicle.ID, image.FileName, image.ContentType,
			image.Width, image.Height, image.SizeBytes, index, image.Data,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (p *Postgres) VehicleByID(id string) (domain.Vehicle, bool) {
	vehicle, err := scanVehicle(p.pool.QueryRow(context.Background(), vehicleSelect+` WHERE id = $1`, id))
	if err != nil {
		return domain.Vehicle{}, false
	}
	images, err := p.imagesByVehicleIDs(context.Background(), []string{id})
	if err != nil {
		return domain.Vehicle{}, false
	}
	vehicle.Images = images[id]
	return vehicle, true
}

func (p *Postgres) VehiclesByStatus(status domain.VehicleStatus) []domain.Vehicle {
	return p.queryVehicles(`WHERE status = $1`, string(status))
}

func (p *Postgres) VehiclesBySeller(sellerID string) []domain.Vehicle {
	return p.queryVehicles(`WHERE seller_id = $1`, sellerID)
}

func (p *Postgres) AllVehicles() []domain.Vehicle {
	return p.queryVehicles("")
}

func (p *Postgres) ApproveVehicle(id string) error {
	return p.changeStatus(id, `UPDATE vehicles SET status = $2 WHERE id = $1`,
		string(domain.VehiclePublished))
}

func (p *Postgres) MarkVehicleSold(id, sellerID string) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status, owner string
	err = tx.QueryRow(ctx, `SELECT status, seller_id FROM vehicles WHERE id = $1 FOR UPDATE`, id).Scan(&status, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if owner != sellerID {
		return ErrNotFound
	}
	if domain.VehicleStatus(status) != domain.VehiclePublished {
		return ErrInvalidState
	}
	if _, err := tx.Exec(ctx, `UPDATE vehicles SET status = $2 WHERE id = $1`, id, string(domain.VehicleSold)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) DeleteVehicle(id string) error {
	tag, err := p.pool.Exec(context.Background(), `DELETE FROM vehicles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) DeleteUser(id string) error {
	tag, err := p.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) changeStatus(id, updateSQL string, args ...any) error {
	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM vehicles WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if domain.VehicleStatus(status) != domain.VehiclePending {
		return ErrInvalidState
	}
	updateArgs := append([]any{id}, args...)
	if _, err := tx.Exec(ctx, updateSQL, updateArgs...); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) queryVehicles(where string, args ...any) []domain.Vehicle {
	ctx := context.Background()
	query := vehicleSelect
	if where != "" {
		query += " " + where
	}
	query += " ORDER BY created_at DESC, id"

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return []domain.Vehicle{}
	}
	defer rows.Close()

	vehicles := make([]domain.Vehicle, 0)
	ids := make([]string, 0)
	for rows.Next() {
		vehicle, err := scanVehicle(rows)
		if err != nil {
			return vehicles
		}
		vehicles = append(vehicles, vehicle)
		ids = append(ids, vehicle.ID)
	}
	images, err := p.imagesByVehicleIDs(ctx, ids)
	if err != nil {
		return vehicles
	}
	for index := range vehicles {
		vehicles[index].Images = images[vehicles[index].ID]
	}
	return vehicles
}

const vehicleSelect = `SELECT id, make, model, year, price_cop, description, seller_id, status, cover_image_id, created_at FROM vehicles`

func (p *Postgres) imagesByVehicleIDs(ctx context.Context, ids []string) (map[string][]domain.VehicleImage, error) {
	result := make(map[string][]domain.VehicleImage)
	if len(ids) == 0 {
		return result, nil
	}
	rows, err := p.pool.Query(ctx, `
		SELECT vehicle_id, id, file_name, content_type, width, height, size_bytes
		FROM vehicle_images WHERE vehicle_id = ANY($1)
		ORDER BY vehicle_id, position, id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var vehicleID string
		var image domain.VehicleImage
		if err := rows.Scan(&vehicleID, &image.ID, &image.FileName, &image.ContentType,
			&image.Width, &image.Height, &image.SizeBytes); err != nil {
			return nil, err
		}
		result[vehicleID] = append(result[vehicleID], image)
	}
	return result, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (domain.User, error) {
	var user domain.User
	var role string
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &role,
		&user.PasswordSalt, &user.PasswordHash, &user.CreatedAt)
	user.Role = domain.Role(role)
	return user, err
}

func scanVehicle(row rowScanner) (domain.Vehicle, error) {
	var vehicle domain.Vehicle
	var status string
	err := row.Scan(&vehicle.ID, &vehicle.Make, &vehicle.Model, &vehicle.Year, &vehicle.PriceCOP,
		&vehicle.Description, &vehicle.SellerID, &status, &vehicle.CoverImageID,
		&vehicle.CreatedAt)
	vehicle.Status = domain.VehicleStatus(status)
	return vehicle, err
}
