package handler

import (
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
)

// BlogsIndex renders /blogs with minified list (no images)
func (h *Handler) BlogsIndex(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	posts, err := h.BlogService.ListPublished(r.Context(), 20, 0)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	// Enrich with reading time and formatted date
	type postView struct {
		Title       string
		Slug        string
		Excerpt     string
		CreatedAt   time.Time
		ReadingTime int
	}
	items := make([]postView, 0, len(posts))
	for _, p := range posts {
		words := 0
		for _, ch := range p.ContentMD {
			if ch == ' ' || ch == '\n' || ch == '\t' {
				words++
			}
		}
		// Minimum 1 minute read
		rt := words / 200
		if rt < 1 {
			rt = 1
		}
		items = append(items, postView{
			Title:       p.Title,
			Slug:        p.Slug,
			Excerpt:     p.Excerpt,
			CreatedAt:   p.CreatedAt,
			ReadingTime: rt,
		})
	}

	t, err := template.New("blogs-index").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/blog/index.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	meta := map[string]any{
		"Title":       "Блог — justlaunch",
		"Description": "Новости, гайды и обновления платформы.",
		"Canonical":   h.BaseURL + "/blogs",
	}

	if err := t.ExecuteTemplate(w, "layout", map[string]any{
		"User":  u,
		"Posts": items,
		"token": nosurf.Token(r),
		"meta":  meta,
	}); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

// BlogShow renders a blog post page; supports embedding product or launch cards via simple shortcodes in content.
// Supported tokens in content (one of each is supported for now):
// [[product:product-slug]]
// [[launch:product-slug:index]]
func (h *Handler) BlogShow(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	slug := chi.URLParam(r, "slug")
	post, err := h.BlogService.GetBySlug(r.Context(), slug)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	var embeddedProduct *product.Product
	var launchItem map[string]any
	var awardsByLaunch map[uuid.UUID][]*launch.LaunchAward

	// Detect product token
	if m := regexp.MustCompile(`\[\[product:([a-z0-9-]+)\]\]`).FindStringSubmatch(post.ContentMD); len(m) == 2 {
		prodSlug := m[1]
		p, err := h.ProductService.GetBySlug(r.Context(), prodSlug)
		if err == nil && p != nil && !p.IsNil() {
			embeddedProduct = p
		}
	}

	// Detect launch token
	if m := regexp.MustCompile(`\[\[launch:([a-z0-9-]+):(\d+)\]\]`).FindStringSubmatch(post.ContentMD); len(m) == 3 {
		prodSlug := m[1]
		idx, _ := strconv.Atoi(m[2])
		p, err := h.ProductService.GetBySlug(r.Context(), prodSlug)
		if err == nil && p != nil && !p.IsNil() {
			l, err := h.LaunchService.GetNthByProductOrderedByCreatedAt(r.Context(), p.ID, idx)
			if err == nil && l != nil {
				// Fetch awards for this launch to satisfy launch-card.html expectations
				if m, err := h.LaunchService.GetAwardsByLaunchIDs(r.Context(), []uuid.UUID{l.ID}); err == nil {
					awardsByLaunch = m
				}
				launchItem = map[string]any{
					"Launch":         l,
					"ProductSlug":    prodSlug,
					"Index":          idx,
					"Upvoted":        false,
					"AwardsByLaunch": awardsByLaunch,
				}
			}
		}
	}

	t, err := template.New("blog-show").Funcs(template.FuncMap{
		"dict":     tmpl.Dict,
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).ParseFiles(
		"views/blog/show.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/product-card.html",
		"views/partials/launch-card.html",
		"views/partials/launch-upvote.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	meta := map[string]any{
		"Title":     post.Title + " — блог justlaunch",
		"Canonical": h.BaseURL + "/blogs/" + post.Slug,
	}

	// Remove embed tokens from the content to avoid leaking raw shortcodes
	productToken := regexp.MustCompile(`\[\[product:([a-z0-9-]+)\]\]`)
	launchToken := regexp.MustCompile(`\[\[launch:([a-z0-9-]+):(\d+)\]\]`)
	cleanContent := productToken.ReplaceAllString(post.ContentMD, "")
	cleanContent = launchToken.ReplaceAllString(cleanContent, "")

	if err := t.ExecuteTemplate(w, "layout", map[string]any{
		"User":           u,
		"Post":           post,
		"PostHTML":       cleanContent,
		"Product":        embeddedProduct,
		"LaunchItem":     launchItem,
		"AwardsByLaunch": awardsByLaunch,
		"token":          nosurf.Token(r),
		"meta":           meta,
	}); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}
