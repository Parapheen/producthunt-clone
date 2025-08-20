package app

import (
	"bytes"
	"context"
	"errors"

	"github.com/Parapheen/ph-clone/internal/domain/blog"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

type BlogService struct {
	repo      blog.PostRepository
	sanitizer *bluemonday.Policy
}

func NewBlogService(repo blog.PostRepository) *BlogService {
	return &BlogService{
		repo: repo,
		sanitizer: func() *bluemonday.Policy {
			p := bluemonday.UGCPolicy()
			p.AllowAttrs("href").OnElements("a")
			p.AllowAttrs("target").OnElements("a")
			p.AllowAttrs("rel").OnElements("a")
			return p
		}(),
	}
}

func (s *BlogService) Create(ctx context.Context, p *blog.Post) error {
	if p == nil {
		return errors.New("nil post")
	}
	if err := p.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, p)
}

func (s *BlogService) Update(ctx context.Context, p *blog.Post) error {
	if p == nil {
		return errors.New("nil post")
	}
	if err := p.Validate(); err != nil {
		return err
	}
	p.Touch()
	return s.repo.Update(ctx, p)
}

func (s *BlogService) GetBySlug(ctx context.Context, slug string) (*blog.Post, error) {
    p, err := s.repo.GetBySlug(ctx, slug)
    if err != nil {
        return nil, err
    }
    var buf bytes.Buffer
    if err := goldmark.Convert([]byte(p.ContentMD), &buf); err == nil {
        safe := s.sanitizer.SanitizeBytes(buf.Bytes())
        p.ContentMD = string(safe)
    } else {
        p.ContentMD = s.sanitizer.Sanitize(p.ContentMD)
    }
    return p, nil
}

func (s *BlogService) ListPublished(ctx context.Context, limit, offset int) ([]*blog.Post, error) {
	return s.repo.ListPublished(ctx, limit, offset)
}


