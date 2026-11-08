package audit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Event representa un evento de auditoría.
type Event struct {
	Time     time.Time `json:"time"`
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	Status   int       `json:"status"`
	ActorID  string    `json:"actor_id,omitempty"`
	Role     string    `json:"role,omitempty"`
	PrevHash string    `json:"prev_hash"`
	Hash     string    `json:"hash"`
}

type hashContent struct {
	Time     time.Time `json:"time"`
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	Status   int       `json:"status"`
	ActorID  string    `json:"actor_id,omitempty"`
	Role     string    `json:"role,omitempty"`
	PrevHash string    `json:"prev_hash"`
}

// Logger registra eventos de auditoría en la tabla audit_log con cadena de hashes SHA-256.
// El último hash se lee de la base de datos en cada registro, de modo que varias
// instancias de la aplicación que compartan la base mantengan una única cadena.
type Logger struct {
	pool *pgxpool.Pool
}

// NewLogger crea un nuevo logger de auditoría usando PostgreSQL y valida la
// cadena de integridad existente al iniciar.
func NewLogger(pool *pgxpool.Pool) (*Logger, error) {
	if err := validateChainFromPool(pool); err != nil {
		return nil, fmt.Errorf("la cadena de integridad de auditoría no es válida: %w", err)
	}
	return &Logger{pool: pool}, nil
}

// validateChainFromPool valida la cadena de hashes SHA-256 en la tabla audit_log.
// Lee eventos en streaming por id ASC, verificando:
// - El primer evento debe tener prev_hash vacío
// - Cada prev_hash debe ser igual al hash del evento anterior
// - hash debe ser igual a hashEvent(event)
func validateChainFromPool(pool *pgxpool.Pool) error {
	var count int
	err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM audit_log`).Scan(&count)
	if err != nil {
		return fmt.Errorf("contar eventos: %w", err)
	}

	rows, err := pool.Query(context.Background(), `
		SELECT id, time, method, path, status, actor_id, role, prev_hash, hash
		FROM audit_log ORDER BY id ASC`)
	if err != nil {
		return fmt.Errorf("obtener eventos: %w", err)
	}
	defer rows.Close()

	var previousHash string
	firstEvent := true

	for rows.Next() {
		var id int
		var event Event
		var actorID sql.NullString
		var role sql.NullString
		if err := rows.Scan(&id, &event.Time, &event.Method, &event.Path, &event.Status,
			&actorID, &role, &event.PrevHash, &event.Hash); err != nil {

			return fmt.Errorf("escaneo evento ID %d: %w", id, err)
		}

		// Asignar ActorID y Role desde sql.NullString después del Scan
		event.ActorID = actorID.String
		event.Role = role.String

		if firstEvent {
			// El primer evento debe tener prev_hash vacío
			if event.PrevHash != "" {
				return fmt.Errorf("primer evento ID %d: prev_hash debe estar vacío", id)
			}
			firstEvent = false
		} else {
			// Cada prev_hash debe ser igual al hash del evento anterior
			if event.PrevHash != previousHash {
				return fmt.Errorf("cadena rota en evento ID %d: prev_hash no coincide", id)
			}
		}

		// Actualizar previousHash para el siguiente evento
		previousHash = event.Hash

		// Calcular el hash esperado (con Time truncado a UTC)
		event.Time = event.Time.UTC()
		expectedHash := hashEvent(event)
		if event.Hash != expectedHash {
			return fmt.Errorf("hash inválido en evento ID %d: calculado=%s, almacenado=%s",
				id, expectedHash, event.Hash)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error después de escanear: %w", err)
	}

	return nil
}

// hashEvent calcula el hash SHA-256 de un evento.
// La función debe producir el mismo hash que se usaba en el archivo data/audit.jsonl
// para conservar la integridad del Accounting. Usa ActorID y Role como cadenas
// vacías (no NULL) al reconstruir el hash para que coincida con el almacenado.
func hashEvent(event Event) string {
	// Time ya debe estar truncado a UTC en este punto
	content, _ := json.Marshal(hashContent{
		Time:     event.Time,
		Method:   event.Method,
		Path:     event.Path,
		Status:   event.Status,
		ActorID:  event.ActorID,
		Role:     event.Role,
		PrevHash: event.PrevHash,
	})
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// Record inserta un evento nuevo en la tabla audit_log, manteniendo la cadena de hashes.
// Un bloqueo asesor de transacción serializa los registros incluso entre instancias:
// el prev_hash se lee dentro de la misma transacción que inserta el evento.
func (l *Logger) Record(event Event) error {
	ctx := context.Background()
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7301)`); err != nil {
		return fmt.Errorf("bloquear cadena de auditoría: %w", err)
	}

	var previousHash string
	err = tx.QueryRow(ctx, `SELECT hash FROM audit_log ORDER BY id DESC LIMIT 1`).Scan(&previousHash)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("obtener último hash: %w", err)
	}

	event.Time = event.Time.UTC().Truncate(time.Microsecond)
	event.PrevHash = previousHash
	event.Hash = hashEvent(event)

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_log (time, method, path, status, actor_id, role, prev_hash, hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.Time, event.Method, event.Path, event.Status,
		event.ActorID, event.Role, event.PrevHash, event.Hash); err != nil {
		return fmt.Errorf("insertar evento auditoría: %w", err)
	}
	return tx.Commit(ctx)
}
