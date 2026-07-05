package modelstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxHuggingFaceModelSize = 30 << 30 // 30 GiB

var (
	errInvalidHuggingFaceRef = errors.New("invalid Hugging Face model ref")
	huggingFaceHTTPClient    = &http.Client{Timeout: 30 * time.Minute}
)

// DownloadHuggingFaceModel downloads a GGUF model from Hugging Face into modelsDir.
// Ref supports either:
//   - full URL (for example https://huggingface.co/org/repo/resolve/main/model.gguf)
//   - shorthand (for example org/repo/path/model.gguf)
func DownloadHuggingFaceModel(ctx context.Context, modelsDir, ref string) (string, error) {
	downloadURL, fileName, err := resolveHuggingFaceRef(ref)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		return "", fmt.Errorf("create models dir: %w", err)
	}
	targetPath := filepath.Join(modelsDir, fileName)
	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	resp, err := huggingFaceHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download model: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download model: unexpected status %s", resp.Status)
	}
	if resp.ContentLength > maxHuggingFaceModelSize {
		return "", fmt.Errorf("download model: file too large (%d bytes)", resp.ContentLength)
	}
	tmpFile, err := os.CreateTemp(modelsDir, "*.gguf.part")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()
	written, err := io.Copy(tmpFile, io.LimitReader(resp.Body, maxHuggingFaceModelSize+1))
	if err != nil {
		return "", fmt.Errorf("write model: %w", err)
	}
	if written > maxHuggingFaceModelSize {
		return "", fmt.Errorf("download model: file too large (%d bytes)", written)
	}
	if err := tmpFile.Sync(); err != nil {
		return "", fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return "", fmt.Errorf("persist model: %w", err)
	}
	return targetPath, nil
}

func resolveHuggingFaceRef(ref string) (downloadURL, fileName string, err error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", "", errInvalidHuggingFaceRef
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		parsed, parseErr := url.Parse(trimmed)
		if parseErr != nil {
			return "", "", fmt.Errorf("%w: %s", errInvalidHuggingFaceRef, ref)
		}
		name := path.Base(parsed.Path)
		if !strings.EqualFold(path.Ext(name), ".gguf") {
			return "", "", fmt.Errorf("%w: expected .gguf file", errInvalidHuggingFaceRef)
		}
		return parsed.String(), name, nil
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("%w: expected owner/repo/path.gguf", errInvalidHuggingFaceRef)
	}
	fileName = parts[len(parts)-1]
	if !strings.EqualFold(filepath.Ext(fileName), ".gguf") {
		return "", "", fmt.Errorf("%w: expected .gguf file", errInvalidHuggingFaceRef)
	}
	repo := strings.Join(parts[:2], "/")
	modelPath := strings.Join(parts[2:], "/")
	return fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", repo, modelPath), fileName, nil
}
