package handler

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (s *Handler) Home(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

    period := r.URL.Query().Get("period")
    if period == "" {
        period = "daily"
    }

    launches, err := s.LaunchService.GetFeedByPeriod(r.Context(), period)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    // Build items with product categories for each launch
    // 1) collect product IDs
    productIDSet := map[uuid.UUID]struct{}{}
    for _, l := range launches {
        productIDSet[l.ProductID] = struct{}{}
    }
    productIDs := make([]uuid.UUID, 0, len(productIDSet))
    for id := range productIDSet {
        productIDs = append(productIDs, id)
    }
    // 2) fetch products with categories in bulk
    products, err := s.ProductService.GetByIDs(r.Context(), productIDs)
    if err != nil {
        s.Logger.ErrorContext(r.Context(), "error getting products for launches", slog.Any("error", err))
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    categoriesByProduct := map[uuid.UUID]interface{}{}
    for _, p := range products {
        categoriesByProduct[p.ID] = p.Categories
    }
    // 3) Upvoted map for user
    // Upvoted map for user when service supports it
    var upvoted map[uuid.UUID]bool
    if user != nil {
        type upvoteAware interface { GetUpvotedMap(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) }
        if svc, ok := any(s.LaunchService).(upvoteAware); ok {
            ids := make([]uuid.UUID, 0, len(launches))
            for _, l := range launches { ids = append(ids, l.ID) }
            up, _ := svc.GetUpvotedMap(r.Context(), user.ID, ids)
            upvoted = up
        }
    }
    // 4) compose items for template
    items := make([]map[string]interface{}, 0, len(launches))
    for _, l := range launches {
        items = append(items, map[string]interface{}{
            "Launch":      l,
            "Categories":  categoriesByProduct[l.ProductID],
            "Upvoted":     upvoted[l.ID],
        })
    }

    // HTMX partial render for feed swap
    if r.Header.Get("HX-Request") == "true" {
        t, err := template.New("home-feed").Funcs(template.FuncMap{
            "dict": tmpl.Dict,
        }).ParseFiles(
            "views/partials/home-feed.html",
            "views/partials/launch-card.html",
            "views/partials/launch-state.html",
            "views/partials/launch-upvote.html",
        )
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        err = t.ExecuteTemplate(w, "home-feed", map[string]interface{}{
            "User":         user,
            "Items":        items,
            "ActivePeriod": period,
            "token":        nosurf.Token(r),
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    t, err := template.New("home-index").Funcs(template.FuncMap{
        "dict": tmpl.Dict,
    }).ParseFiles(
        "views/index.html",
        "views/layout/layout.html",
        "views/layout/header.html",
        "views/layout/footer.html",
        "views/layout/head.html",
        "views/partials/home-feed.html",
        "views/partials/launch-card.html",
        "views/partials/launch-state.html",
        "views/partials/launch-upvote.html",
    )
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
        "User":         user,
        "Items":        items,
        "ActivePeriod": period,
        "token":        nosurf.Token(r),
    })
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
