-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL CHECK (role IN ('seller', 'admin')),
    password_salt TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_idx ON users (LOWER(email));

-- Tabla de vehículos
CREATE TABLE IF NOT EXISTS vehicles (
    id TEXT PRIMARY KEY,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER NOT NULL,
    price_cop BIGINT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    seller_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'published', 'sold')),
    cover_image_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS vehicles_status_created_idx ON vehicles (status, created_at DESC);
CREATE INDEX IF NOT EXISTS vehicles_seller_created_idx ON vehicles (seller_id, created_at DESC);

-- Tabla de imágenes de vehículos (con datos binarios en PostgreSQL)
CREATE TABLE IF NOT EXISTS vehicle_images (
    id TEXT PRIMARY KEY,
    vehicle_id TEXT NOT NULL REFERENCES vehicles (id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    size_bytes BIGINT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    data BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS vehicle_images_vehicle_idx ON vehicle_images (vehicle_id, position);

-- Para bases de datos existentes: agregar columna data si no existe
ALTER TABLE vehicle_images ADD COLUMN IF NOT EXISTS data BYTEA;

-- Tabla de auditoría con cadena de hashes SHA-256
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    time TIMESTAMPTZ NOT NULL,
    method TEXT NOT NULL,
    path TEXT NOT NULL,
    status INTEGER NOT NULL,
    actor_id TEXT,
    role TEXT,
    prev_hash TEXT NOT NULL,
    hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS audit_log_time_idx ON audit_log (id DESC);