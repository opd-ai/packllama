package modelstore

import (
	"os"
	"path/filepath"
	"testing"
)

// makeGGUF creates a zero-byte .gguf file at the given path.
func makeGGUF(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	f.Close()
}

func TestScan_FindsGGUFFiles(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "llama-7b.gguf"))
	makeGGUF(t, filepath.Join(dir, "mistral-7b.gguf"))
	// Non-GGUF files should be ignored.
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore"), 0o644)

	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	entries := r.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestScan_EmptyDir(t *testing.T) {
	r := New()
	if err := r.Scan(t.TempDir(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.List()) != 0 {
		t.Error("expected empty list for empty dir")
	}
}

func TestScan_MissingDir(t *testing.T) {
	r := New()
	// Non-existent directory should not error.
	if err := r.Scan("/nonexistent/dir", false); err != nil {
		t.Fatalf("expected no error for missing dir, got: %v", err)
	}
}

func TestScan_EmptyModelsDir(t *testing.T) {
	r := New()
	// Empty string means "not configured" — no error.
	if err := r.Scan("", false); err != nil {
		t.Fatalf("unexpected error for empty dir: %v", err)
	}
}

func TestScan_Recursive(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "top.gguf"))
	makeGGUF(t, filepath.Join(dir, "sub", "nested.gguf"))

	r := New()
	if err := r.Scan(dir, true); err != nil {
		t.Fatalf("Scan recursive: %v", err)
	}
	if len(r.List()) != 2 {
		t.Fatalf("expected 2 entries with recursive scan, got %d", len(r.List()))
	}
}

func TestScan_NonRecursive(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "top.gguf"))
	makeGGUF(t, filepath.Join(dir, "sub", "nested.gguf"))

	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan non-recursive: %v", err)
	}
	if len(r.List()) != 1 {
		t.Fatalf("expected 1 entry without recursion, got %d", len(r.List()))
	}
}

func TestGet_Found(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "llama-7b.gguf"))

	r := New()
	r.Scan(dir, false)

	e, ok := r.Get("llama-7b")
	if !ok {
		t.Fatal("expected to find llama-7b")
	}
	if e.ID != "llama-7b" {
		t.Errorf("unexpected ID %q", e.ID)
	}
	if e.OwnedBy != "local" {
		t.Errorf("unexpected OwnedBy %q", e.OwnedBy)
	}
}

func TestGet_NotFound(t *testing.T) {
	r := New()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestAlias(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "llama-7b.gguf"))

	r := New()
	r.Scan(dir, false)
	r.AddAlias("default", "llama-7b")

	e, ok := r.Get("default")
	if !ok {
		t.Fatal("expected alias to resolve")
	}
	if e.ID != "llama-7b" {
		t.Errorf("alias resolved to %q, expected llama-7b", e.ID)
	}
}

func TestResolve(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "model.gguf"))

	r := New()
	r.Scan(dir, false)

	path, err := r.Resolve("model")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestResolve_NotFound(t *testing.T) {
	r := New()
	if _, err := r.Resolve("missing"); err == nil {
		t.Error("expected error for missing model")
	}
}
