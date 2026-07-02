// Package modelstore provides model discovery and registry for packllama.
// It scans configurable directories for GGUF model files and maintains a
// cached registry of model metadata.
package modelstore

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Entry describes a single discovered model file.
type Entry struct {
	// ID is the canonical model identifier derived from the file name without extension.
	ID string
	// Path is the absolute path to the .gguf file.
	Path string
	// Size is the file size in bytes.
	Size int64
	// ModTime is the file modification time.
	ModTime time.Time
	// OwnedBy is the owner label; defaults to "local".
	OwnedBy string
}

// Registry holds a cached list of discovered models and supports aliases.
type Registry struct {
	mu      sync.RWMutex
	entries []Entry
	aliases map[string]string // alias → model ID
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		aliases: make(map[string]string),
	}
}

// Scan walks dir and registers every *.gguf file it finds. Existing entries
// are replaced by the result of this scan. The scan is shallow by default;
// pass recursive=true to descend into subdirectories.
func (r *Registry) Scan(dir string, recursive bool) error {
	if dir == "" {
		return nil
	}
	var entries []Entry
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if !recursive && path != dir {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".gguf") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("abs path %s: %w", path, err)
		}
		id := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		entries = append(entries, Entry{
			ID:      id,
			Path:    abs,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			OwnedBy: "local",
		})
		return nil
	}
	if err := filepath.WalkDir(dir, walkFn); err != nil {
		if os.IsNotExist(err) {
			return nil // dir not yet created is not an error
		}
		return fmt.Errorf("scan %s: %w", dir, err)
	}
	r.mu.Lock()
	r.entries = entries
	r.mu.Unlock()
	return nil
}

// AddAlias registers an alias that maps to an existing model ID.
// It returns false if modelID does not match any registered entry.
func (r *Registry) AddAlias(alias, modelID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	found := false
	for _, e := range r.entries {
		if e.ID == modelID {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	r.aliases[alias] = modelID
	return true
}

// List returns all registered model entries.
func (r *Registry) List() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Entry, len(r.entries))
	copy(out, r.entries)
	return out
}

// Get looks up a model by ID or alias. It returns the entry and true when found.
func (r *Registry) Get(id string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Resolve alias.
	if resolved, ok := r.aliases[id]; ok {
		id = resolved
	}
	for _, e := range r.entries {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

// Resolve returns the file path for the given model ID or alias.
func (r *Registry) Resolve(id string) (string, error) {
	e, ok := r.Get(id)
	if !ok {
		return "", fmt.Errorf("model %q not found", id)
	}
	return e.Path, nil
}
