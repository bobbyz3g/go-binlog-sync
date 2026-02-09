package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	Path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{Path: path}
}

func (s *FileStore) Load(ctx context.Context) (*State, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s == nil || s.Path == "" {
		return nil, errors.New("state file path is empty")
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("decode state file: %w", err)
	}
	return &st, nil
}

func (s *FileStore) Save(ctx context.Context, st *State) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.Path == "" {
		return errors.New("state file path is empty")
	}
	if st == nil {
		return errors.New("state is nil")
	}
	st.Touch(timeNowUTC())

	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	file, err := os.CreateTemp(dir, ".gbs-state-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp state file: %w", err)
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()

	enc := json.NewEncoder(file)
	if err := enc.Encode(st); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode state file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync state file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close state file: %w", err)
	}
	if err := os.Rename(file.Name(), s.Path); err != nil {
		return fmt.Errorf("rename state file: %w", err)
	}
	return nil
}
