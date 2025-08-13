package handler

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

// GetLaunch renders a standalone launch page by slug (legacy). Redirect to nested URL.
func (h *Handler) GetLaunch(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("launchSlug")

    l, err := h.LaunchService.GetBySlug(r.Context(), slug)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Load product for context and permissions
    p, err := h.ProductService.GetByID(r.Context(), l.ProductID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Permanent redirect to nested route
    http.Redirect(w, r, "/products/"+p.Slug+"/launches/"+l.Slug, http.StatusMovedPermanently)
}

// GetProductLaunch renders a launch page nested under the product slug
func (h *Handler) GetProductLaunch(w http.ResponseWriter, r *http.Request) {
    u := user.GetUserFromContext(r.Context())

    productSlug := r.PathValue("productSlug")
    launchSlug := r.PathValue("launchSlug")

    p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    l, err := h.LaunchService.GetByProductAndSlug(r.Context(), p.ID, launchSlug)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Determine if current user upvoted this launch
    var upvoted bool
    if u != nil {
        type upvoteAware interface {
            GetUpvotedMap(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error)
        }
        if svc, ok := any(h.LaunchService).(upvoteAware); ok {
            m, _ := svc.GetUpvotedMap(r.Context(), u.ID, []uuid.UUID{l.ID})
            upvoted = m[l.ID]
        }
    }

    // Helpers
    humanTime := func(ts time.Time) string {
        d := time.Since(ts)
        if d < time.Minute {
            return "только что"
        }
        if d < time.Hour {
            return strconv.Itoa(int(d.Minutes())) + " мин назад"
        }
        if d < 24*time.Hour {
            return strconv.Itoa(int(d.Hours())) + " ч назад"
        }
        days := int(d.Hours() / 24)
        if days < 30 {
            return strconv.Itoa(days) + " дн назад"
        }
        months := days / 30
        if months < 12 {
            return strconv.Itoa(months) + " мес назад"
        }
        years := months / 12
        return strconv.Itoa(years) + " г назад"
    }

    t, err := template.New("launch.html").Funcs(template.FuncMap{
        "dict":          tmpl.Dict,
        "safeHTML":      func(s string) template.HTML { return template.HTML(s) },
        "humanTime":     humanTime,
        "formatDateTime": func(ts time.Time) string { return ts.Format("02.01.2006 15:04") },
    }).ParseFiles(
        "views/launch.html",
        "views/layout/layout.html",
        "views/layout/header.html",
        "views/layout/footer.html",
        "views/layout/head.html",
        "views/partials/launch-upvote.html",
        "views/partials/launch-comments.html",
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // SEO meta for Launch
    canonical := h.BaseURL + "/products/" + p.Slug + "/launches/" + l.Slug
    meta := map[string]any{
        "Title":       l.Name + " — запуск продукта " + p.Name,
        "Description": l.Tagline,
        "Canonical":   canonical,
        "OGType":      "article",
        "Image":       l.ImageURL,
    }

    err = t.ExecuteTemplate(w, "layout", map[string]any{
        "User":     u,
        "Product":  p,
        "Launch":   l,
        "Upvoted":  upvoted,
        "token":    nosurf.Token(r),
        "meta":     meta,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

