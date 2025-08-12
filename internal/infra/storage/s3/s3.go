package s3

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parapheen/ph-clone/internal/app"
)

// Minimal interface so we don't couple to a specific SDK here; wire real client in main if needed.
type Uploader interface {
    PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) (publicURL string, err error)
    DeleteObject(ctx context.Context, bucket, key string) error
}

type S3Storage struct {
    bucket string
    baseKey string
    uploader Uploader
}

func NewS3Storage(bucket, baseKey string, uploader Uploader) *S3Storage {
    return &S3Storage{bucket: bucket, baseKey: strings.Trim(baseKey, "/"), uploader: uploader}
}

var _ app.Storage = (*S3Storage)(nil)

func (s *S3Storage) Save(ctx context.Context, pathPrefix string, originalFilename string, r io.Reader) (string, error) {
    safePrefix := sanitize(pathPrefix)
    if safePrefix == "" {
        safePrefix = "misc"
    }
    ext := filepath.Ext(originalFilename)
    if ext == "" {
        ext = ".bin"
    }
    h := sha1.New()
    h.Write([]byte(fmt.Sprintf("%s-%d-%s", originalFilename, time.Now().UnixNano(), safePrefix)))
    filename := hex.EncodeToString(h.Sum(nil)) + ext
    key := strings.Trim(s.baseKey, "/") + "/" + safePrefix + "/" + filename

    contentType := mime.TypeByExtension(ext)
    url, err := s.uploader.PutObject(ctx, s.bucket, key, r, contentType)
    if err != nil {
        return "", err
    }
    return url, nil
}

func (s *S3Storage) Delete(ctx context.Context, publicURL string) error {
    // If uploader can map URL->key, implement here. Otherwise, best-effort no-op.
    return nil
}

func sanitize(p string) string {
    p = strings.TrimSpace(p)
    p = strings.Trim(p, "/")
    p = strings.ReplaceAll(p, "..", "")
    p = strings.ReplaceAll(p, "\\", "/")
    p = strings.ReplaceAll(p, " ", "-")
    return p
}


