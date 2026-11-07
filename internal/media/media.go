package media

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"

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

// Store guarda las imágenes en disco, fuera del directorio público.
type Store struct {
	root string
}

func New(root string) (*Store, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0o700); err != nil {
		return nil, err
	}
	return &Store{root: absolute}, nil
}

// Save valida que el contenido real del archivo sea una imagen JPG o PNG, no solo su extensión.
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
	directory := filepath.Join(s.root, vehicleID)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return domain.VehicleImage{}, err
	}
	temporary, err := os.CreateTemp(directory, ".upload-*.tmp")
	if err != nil {
		return domain.VehicleImage{}, err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return domain.VehicleImage{}, err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return domain.VehicleImage{}, err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return domain.VehicleImage{}, err
	}
	if err := temporary.Close(); err != nil {
		return domain.VehicleImage{}, err
	}
	if err := os.Rename(temporaryPath, filepath.Join(directory, fileName)); err != nil {
		return domain.VehicleImage{}, err
	}

	return domain.VehicleImage{
		ID:          imageID,
		FileName:    fileName,
		ContentType: contentType,
		Width:       config.Width,
		Height:      config.Height,
		SizeBytes:   int64(len(data)),
	}, nil
}

func (s *Store) DeleteVehicle(vehicleID string) error {
	if !vehicleIDPattern.MatchString(vehicleID) {
		return nil
	}
	err := os.RemoveAll(filepath.Join(s.root, vehicleID))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Resolve devuelve la ruta de una imagen verificando que el nombre sea seguro.
func (s *Store) Resolve(vehicleID, fileName string) (string, bool) {
	if !vehicleIDPattern.MatchString(vehicleID) || !fileNamePattern.MatchString(fileName) {
		return "", false
	}
	path := filepath.Join(s.root, vehicleID, fileName)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", false
	}
	return path, true
}

func newImageID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
