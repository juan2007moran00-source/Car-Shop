package audit

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

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

type Logger struct {
	mu       sync.Mutex
	path     string
	lastHash string
}

func NewLogger(path string) (*Logger, error) {
	logger := &Logger{path: path}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return logger, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, errors.New("el archivo de auditoría contiene una línea inválida")
		}
		if event.PrevHash != logger.lastHash || event.Hash != hashEvent(event) {
			return nil, errors.New("la cadena de integridad de auditoría no es válida")
		}
		logger.lastHash = event.Hash
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return logger, nil
}

func (l *Logger) Record(event Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	event.Time = event.Time.UTC()
	event.PrevHash = l.lastHash
	event.Hash = hashEvent(event)
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	l.lastHash = event.Hash
	return nil
}

func hashEvent(event Event) string {
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
