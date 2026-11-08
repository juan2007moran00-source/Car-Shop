package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/juan2007moran00-source/Car-Shop/internal/domain"
)

var (
	ErrInvalidImage     = errors.New("la imagen no es válida")
	ErrUnsupportedImage = errors.New("solo se permiten imágenes JPG o PNG")
)

var (
	vehicleIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)
	fileNamePattern  = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png)$`)
)

// Store gestiona el almacenamiento de imágenes en PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// New crea un nuevo almacenamiento de imágenes usando PostgreSQL.
func New(pool *pgxpool.Pool) (*Store, error) {
	if pool == nil {
		return nil, errors.New("pool de PostgreSQL nulo")
	}
	return &Store{pool: pool}, nil
}

// Save valida que el contenido real del archivo sea una imagen JPG o PNG,
// y devuelve los bytes en domain.VehicleImage con el campo Data poblado.
// NO inserta nada en la base de datos.
func (s *Store) Save(vehicleID string, reader io.Reader) (domain.VehicleImage, error) {
	if !vehicleIDPattern.MatchString(vehicleID) {
		return domain.VehicleImage{}, ErrInvalidImage
	}
	data, err := io.ReadAll(reader)
	if err != nil || len(data) == 0 {
		return domain.VehicleImage{}, ErrInvalidImage
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return domain.VehicleImage{}, ErrInvalidImage
	}

	var extension, contentType string
	switch format {
	case "jpeg":
		extension = "jpg"
		contentType = "image/jpeg"
	case "png":
		extension = "png"
		contentType = "image/png"
	default:
		return domain.VehicleImage{}, ErrUnsupportedImage
	}

	imageID, err := newImageID()
	if err != nil {
		return domain.VehicleImage{}, err
	}
	fileName := imageID + "." + extension

	return domain.VehicleImage{
		ID:          imageID,
		FileName:    fileName,
		ContentType: contentType,
		Width:       config.Width,
		Height:      config.Height,
		SizeBytes:   int64(len(data)),
		Data:        data,
	}, nil
}

// DeleteVehicle ya no es necesario (ON DELETE CASCADE en foreign key).
// Se mantiene por compatibilidad de la interfaz pero es un no-op.
func (s *Store) DeleteVehicle(vehicleID string) error {
	if !vehicleIDPattern.MatchString(vehicleID) {
		return nil
	}
	// ON DELETE CASCADE en la foreign key limpia vehicle_images automáticamente
	return nil
}

// Resolve busca una imagen en la base de datos por vehicle_id y file_name.
// Devuelve el contenido bytes y el Content-Type.
func (s *Store) Resolve(vehicleID, fileName string) ([]byte, string, bool) {
	if !vehicleIDPattern.MatchString(vehicleID) || !fileNamePattern.MatchString(fileName) {
		return nil, "", false
	}

	var data []byte
	var contentType string
	err := s.pool.QueryRow(context.Background(), `
		SELECT data, content_type FROM vehicle_images WHERE vehicle_id = $1 AND file_name = $2`,
		vehicleID, fileName).Scan(&data, &contentType)
	if err != nil {
		return nil, "", false
	}
	if data == nil {
		return nil, "", false
	}
	return data, contentType, true
}

func newImageID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
