package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var ErrNotFound = errors.New("store: record not found")

type Store struct {
	mu   sync.Mutex
	root string
}

func New(dir string) *Store {
	return &Store{root: dir}
}

func (s *Store) Dir() string {
	return s.root
}

func (s *Store) WriteJSON(rel string, v any) error {
	path := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("store: mkdir %s: %w", rel, err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("store: marshal %s: %w", rel, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("store: write %s: %w", rel, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("store: commit %s: %w", rel, err)
	}
	return nil
}

func (s *Store) ReadJSON(rel string, v any) error {
	path := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("store: read %s: %w", rel, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("store: unmarshal %s: %w", rel, err)
	}
	return nil
}

func (s *Store) Delete(rel string) error {
	path := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("store: delete %s: %w", rel, err)
	}
	return nil
}

func (s *Store) Exists(rel string) bool {
	path := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := os.Stat(path)
	return err == nil
}

func (s *Store) AppendLine(rel, line string) error {
	path := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("store: mkdir %s: %w", rel, err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("store: append %s: %w", rel, err)
	}
	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("store: write %s: %w", rel, err)
	}
	return nil
}

func (s *Store) ReadLines(rel string) ([]string, error) {
	path := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("store: read %s: %w", rel, err)
	}
	var lines []string
	for _, raw := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(raw) != "" {
			lines = append(lines, raw)
		}
	}
	return lines, nil
}

func (s *Store) ListFiles(rel string) ([]string, error) {
	dir := filepath.Join(s.root, filepath.FromSlash(rel))
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("store: list %s: %w", rel, err)
	}
	var out []string
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}
		out = append(out, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
	}
	sort.Strings(out)
	return out, nil
}
