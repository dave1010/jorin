package session

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dave1010/jorin/internal/types"
)

// Store represents a persistence backend for chat sessions.
type Store interface {
	Save(id string, msgs []types.Message) error
	Load(id string) ([]types.Message, error)
	List() ([]string, error)
	Delete(id string) error
}

// FileStore is a simple file-backed Store implementation that writes one
// JSON file per session under a base directory.
type FileStore struct {
	BaseDir string
}

func NewFileStore(baseDir string) *FileStore {
	return &FileStore{BaseDir: baseDir}
}

func (f *FileStore) pathFor(id string) string {
	return filepath.Join(f.BaseDir, id+".json")
}

func (f *FileStore) Save(id string, msgs []types.Message) error {
	if id == "" {
		return errors.New("missing id")
	}
	if err := os.MkdirAll(f.BaseDir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.pathFor(id), b, 0o644)
}

func (f *FileStore) Load(id string) ([]types.Message, error) {
	if id == "" {
		return nil, errors.New("missing id")
	}
	b, err := os.ReadFile(f.pathFor(id))
	if err != nil {
		return nil, err
	}
	var msgs []types.Message
	if err := json.Unmarshal(b, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

func (f *FileStore) List() ([]string, error) {
	ents, err := os.ReadDir(f.BaseDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	out := []string{}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".json" {
			out = append(out, name[:len(name)-len(".json")])
		}
	}
	return out, nil
}

func (f *FileStore) Delete(id string) error {
	if id == "" {
		return errors.New("missing id")
	}
	return os.Remove(f.pathFor(id))
}
