package local

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parapheen/ph-clone/internal/app"
)

type FilesystemStorage struct {
    // baseDir is an absolute or server-working-directory relative path where files are stored
    baseDir string
    // publicBaseURL is the absolute base URL from which assets are served, e.g. https://example.com/assets/uploads
    publicBaseURL string
}

func NewFilesystemStorage(baseDir, publicBaseURL string) *FilesystemStorage {
    return &FilesystemStorage{baseDir: baseDir, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}
}

// Ensure FilesystemStorage implements app.Storage
var _ app.Storage = (*FilesystemStorage)(nil)

func (fs *FilesystemStorage) Save(ctx context.Context, pathPrefix string, originalFilename string, r io.Reader) (string, error) {
    safePrefix := sanitizePath(pathPrefix)
    if safePrefix == "" {
        safePrefix = "misc"
    }
    // Create directory
    targetDir := filepath.Join(fs.baseDir, safePrefix)
    if err := os.MkdirAll(targetDir, 0o755); err != nil {
        return "", fmt.Errorf("create dir: %w", err)
    }

    // Create a deterministic filename to avoid collisions/leaks
    ext := filepath.Ext(originalFilename)
    if ext == "" {
        ext = ".bin"
    }
    h := sha1.New()
    h.Write([]byte(fmt.Sprintf("%s-%d-%s", originalFilename, time.Now().UnixNano(), safePrefix)))
    filename := hex.EncodeToString(h.Sum(nil)) + ext

    targetPath := filepath.Join(targetDir, filename)
    file, err := os.Create(targetPath)
    if err != nil {
        return "", fmt.Errorf("create file: %w", err)
    }
    defer file.Close()

    if _, err := io.Copy(file, r); err != nil {
        return "", fmt.Errorf("write file: %w", err)
    }

    // Build public URL
    // path under public base should mirror safePrefix/filename
    u, err := url.Parse(fs.publicBaseURL + "/" + safePrefix + "/" + filename)
    if err != nil {
        return "", fmt.Errorf("build public url: %w", err)
    }
    return u.String(), nil
}

func (fs *FilesystemStorage) Delete(ctx context.Context, publicURL string) error {
    // Best-effort: map public URL back to file path within baseDir
    if publicURL == "" {
        return nil
    }
    // Extract the portion after publicBaseURL
    trimmed := strings.TrimPrefix(publicURL, fs.publicBaseURL)
    trimmed = strings.TrimPrefix(trimmed, "/")
    if trimmed == "" {
        return nil
    }
    path := filepath.Join(fs.baseDir, trimmed)
    _ = os.Remove(path)
    return nil
}

func sanitizePath(p string) string {
    p = strings.TrimSpace(p)
    p = strings.Trim(p, "/")
    p = strings.ReplaceAll(p, "..", "")
    p = strings.ReplaceAll(p, "\\", "/")
    p = strings.ReplaceAll(p, " ", "-")
    return p
}


