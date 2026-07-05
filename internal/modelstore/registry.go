// Package modelstore provides model discovery and registry for packllama.
// It scans configurable directories for GGUF model files and maintains a
// cached registry of model metadata.
package modelstore

import (
	"errors"
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

	// Metadata fields — populated when available (e.g. from GGUF header or inference backend).

	// ContextLength is the maximum context window in tokens. Zero when unknown.
	ContextLength int64
	// ParameterCount is the number of model parameters. Zero when unknown.
	ParameterCount int64
	// Quantization describes the weight quantization scheme (e.g. "Q4_K_M"). Empty when unknown.
	Quantization string
}

// Registry holds a cached list of discovered models and supports aliases.
type Registry struct {
	mu      sync.RWMutex
	entries []Entry
	aliases map[string]string // alias → model ID
	roots   []string
}

var (
	// ErrInvalidModelFile is returned when a model path is empty or not a .gguf file.
	ErrInvalidModelFile = errors.New("invalid model file")
	// ErrModelAlreadyExists is returned when adding a model with a duplicate ID.
	ErrModelAlreadyExists = errors.New("model already exists")
	// ErrModelNotFound is returned when removing or resolving a missing model.
	ErrModelNotFound = errors.New("model not found")
)

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
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("abs dir %s: %w", dir, err)
	}
	entries := make([]Entry, 0)
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
	if err := filepath.WalkDir(absDir, walkFn); err != nil {
		if os.IsNotExist(err) {
			r.mu.Lock()
			r.roots = appendUniqueRoot(r.roots, absDir)
			r.entries = entries
			r.mu.Unlock()
			return nil // dir not yet created is not an error
		}
		return fmt.Errorf("scan %s: %w", absDir, err)
	}
	r.mu.Lock()
	r.roots = appendUniqueRoot(r.roots, absDir)
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

// AddModelFile registers a single GGUF model file and returns the added entry.
// If id is empty, it is derived from the file name. ownedBy defaults to "local".
func (r *Registry) AddModelFile(path, id, ownedBy string) (Entry, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || !filepath.IsAbs(cleanPath) || !strings.EqualFold(filepath.Ext(cleanPath), ".gguf") {
		return Entry{}, ErrInvalidModelFile
	}
	r.mu.RLock()
	checkedPath, allowed := canonicalPathWithinRoots(cleanPath, r.roots)
	r.mu.RUnlock()
	if !allowed {
		return Entry{}, ErrInvalidModelFile
	}

	info, err := os.Stat(checkedPath)
	if err != nil {
		return Entry{}, fmt.Errorf("stat %s: %w", checkedPath, err)
	}
	if !info.Mode().IsRegular() {
		return Entry{}, ErrInvalidModelFile
	}
	abs, err := filepath.Abs(checkedPath)
	if err != nil {
		return Entry{}, fmt.Errorf("abs path %s: %w", cleanPath, err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if id == "" {
		id = strings.TrimSuffix(filepath.Base(checkedPath), filepath.Ext(checkedPath))
	}
	if ownedBy == "" {
		ownedBy = "local"
	}
	for _, e := range r.entries {
		if e.ID == id {
			return Entry{}, ErrModelAlreadyExists
		}
	}

	entry := Entry{
		ID:      id,
		Path:    abs,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		OwnedBy: ownedBy,
	}
	r.entries = append(r.entries, entry)
	return entry, nil
}

// RemoveModel unregisters a model by ID.
func (r *Registry) RemoveModel(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, e := range r.entries {
		if e.ID == id {
			r.entries = append(r.entries[:i], r.entries[i+1:]...)
			for alias, target := range r.aliases {
				if target == id {
					delete(r.aliases, alias)
				}
			}
			return nil
		}
	}
	return ErrModelNotFound
}

func appendUniqueRoot(roots []string, root string) []string {
	for _, existing := range roots {
		if existing == root {
			return roots
		}
	}
	return append(roots, root)
}

func canonicalPathWithinRoots(path string, roots []string) (string, bool) {
	for _, root := range roots {
		cleanRoot := filepath.Clean(root)
		if path == cleanRoot {
			continue
		}
		prefix := cleanRoot + string(filepath.Separator)
		if strings.HasPrefix(path, prefix) {
			return path, true
		}
	}
	return "", false
}
