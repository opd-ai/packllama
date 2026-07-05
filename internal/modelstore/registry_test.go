package modelstore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	if ok := r.AddAlias("default", "llama-7b"); !ok {
		t.Fatal("expected AddAlias to return true for existing model")
	}

	e, ok := r.Get("default")
	if !ok {
		t.Fatal("expected alias to resolve")
	}
	if e.ID != "llama-7b" {
		t.Errorf("alias resolved to %q, expected llama-7b", e.ID)
	}
}

func TestAlias_UnknownModel(t *testing.T) {
	r := New()
	if ok := r.AddAlias("alias", "nonexistent"); ok {
		t.Error("expected AddAlias to return false for nonexistent model")
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

func TestAddModelFile(t *testing.T) {
	dir := t.TempDir()
	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	modelPath := filepath.Join(dir, "new.gguf")
	makeGGUF(t, modelPath)
	entry, err := r.AddModelFile(modelPath, "", "")
	if err != nil {
		t.Fatalf("AddModelFile: %v", err)
	}
	if entry.ID != "new" {
		t.Fatalf("expected ID new, got %q", entry.ID)
	}
	if got := len(r.List()); got != 1 {
		t.Fatalf("expected 1 model, got %d", got)
	}
}

func TestAddModelFile_DuplicateID(t *testing.T) {
	dir := t.TempDir()
	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	makeGGUF(t, filepath.Join(dir, "a.gguf"))
	makeGGUF(t, filepath.Join(dir, "b.gguf"))
	if _, err := r.AddModelFile(filepath.Join(dir, "a.gguf"), "dup", ""); err != nil {
		t.Fatalf("first AddModelFile: %v", err)
	}
	if _, err := r.AddModelFile(filepath.Join(dir, "b.gguf"), "dup", ""); !errors.Is(err, ErrModelAlreadyExists) {
		t.Fatalf("expected ErrModelAlreadyExists, got %v", err)
	}
}

func TestAddModelFile_InvalidExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.bin")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if _, err := r.AddModelFile(path, "", ""); !errors.Is(err, ErrInvalidModelFile) {
		t.Fatalf("expected ErrInvalidModelFile, got %v", err)
	}
}

func TestAddModelFile_RejectsRelativePath(t *testing.T) {
	dir := t.TempDir()
	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	modelPath := filepath.Join(dir, "relative.gguf")
	makeGGUF(t, modelPath)
	rel, err := filepath.Rel(dir, modelPath)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}
	if _, err := r.AddModelFile(rel, "", ""); !errors.Is(err, ErrInvalidModelFile) {
		t.Fatalf("expected ErrInvalidModelFile, got %v", err)
	}
}

func TestAddModelFile_RejectsPathOutsideScannedRoots(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	modelPath := filepath.Join(other, "outside.gguf")
	makeGGUF(t, modelPath)

	r := New()
	if err := r.Scan(root, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if _, err := r.AddModelFile(modelPath, "", ""); !errors.Is(err, ErrInvalidModelFile) {
		t.Fatalf("expected ErrInvalidModelFile for path outside root, got %v", err)
	}
}

func TestRemoveModel(t *testing.T) {
	dir := t.TempDir()
	makeGGUF(t, filepath.Join(dir, "a.gguf"))

	r := New()
	if err := r.Scan(dir, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if err := r.RemoveModel("a"); err != nil {
		t.Fatalf("RemoveModel: %v", err)
	}
	if got := len(r.List()); got != 0 {
		t.Fatalf("expected 0 models, got %d", got)
	}
}

func TestRemoveModel_NotFound(t *testing.T) {
	r := New()
	if err := r.RemoveModel("missing"); !errors.Is(err, ErrModelNotFound) {
		t.Fatalf("expected ErrModelNotFound, got %v", err)
	}
}

func TestDownloadHuggingFaceModel_FromURL(t *testing.T) {
	modelData := []byte("gguf-bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/my.gguf" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(modelData)))
		_, _ = w.Write(modelData)
	}))
	defer server.Close()

	dir := t.TempDir()
	path, err := DownloadHuggingFaceModel(context.Background(), dir, server.URL+"/my.gguf")
	if err != nil {
		t.Fatalf("DownloadHuggingFaceModel: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != string(modelData) {
		t.Fatalf("unexpected content %q", string(content))
	}
}

func TestDownloadHuggingFaceModel_InvalidRef(t *testing.T) {
	dir := t.TempDir()
	_, err := DownloadHuggingFaceModel(context.Background(), dir, "org/repo/model.bin")
	if !errors.Is(err, errInvalidHuggingFaceRef) {
		t.Fatalf("expected errInvalidHuggingFaceRef, got %v", err)
	}
}

func TestResolveHuggingFaceRef_Shorthand(t *testing.T) {
	url, fileName, err := resolveHuggingFaceRef("owner/repo/models/model.gguf")
	if err != nil {
		t.Fatalf("resolveHuggingFaceRef: %v", err)
	}
	if fileName != "model.gguf" {
		t.Fatalf("expected model.gguf, got %q", fileName)
	}
	if !strings.Contains(url, "https://huggingface.co/owner/repo/resolve/main/models/model.gguf") {
		t.Fatalf("unexpected URL %q", url)
	}
}
