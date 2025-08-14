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
		s.InternalServerError(w, r, err)
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
        s.InternalServerError(w, r, err)
        return
    }
    categoriesByProduct := map[uuid.UUID]interface{}{}
    productSlugByID := map[uuid.UUID]string{}
    for _, p := range products {
        categoriesByProduct[p.ID] = p.Categories
        productSlugByID[p.ID] = p.Slug
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
    // Build index map per product to avoid per-item SQL; compute by created_at DESC position.
    type indexAware interface { GetIndexByProductAndLaunchID(ctx context.Context, productID, launchID uuid.UUID) (int, error) }
    var idxSvc indexAware
    if svc, ok := any(s.LaunchService).(indexAware); ok { idxSvc = svc }
    for _, l := range launches {
        item := map[string]interface{}{
            "Launch":      l,
            "Categories":  categoriesByProduct[l.ProductID],
            "Upvoted":     upvoted[l.ID],
            "ProductSlug": productSlugByID[l.ProductID],
        }
        if idxSvc != nil {
            if idx, err := idxSvc.GetIndexByProductAndLaunchID(r.Context(), l.ProductID, l.ID); err == nil { item["Index"] = idx }
        }
        items = append(items, item)
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
            s.InternalServerError(w, r, err)
            return
        }

        err = t.ExecuteTemplate(w, "home-feed", map[string]interface{}{
            "User":         user,
            "Items":        items,
            "ActivePeriod": period,
            "token":        nosurf.Token(r),
        })
        if err != nil {
            s.InternalServerError(w, r, err)
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
		s.InternalServerError(w, r, err)
		return
	}

    // SEO meta for Home
    baseURL := s.BaseURL
    meta := map[string]any{
        "Title":       "Лучшие стартапы и запуски — justlaunch",
        "Description": "Открывайте новые продукты, следите за запусками и продвигайте свои проекты. Сообщество создателей и ранних пользователей.",
        "Canonical":   baseURL + "/",
        "OGType":      "website",
    }

    err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
        "User":         user,
        "Items":        items,
        "ActivePeriod": period,
        "token":        nosurf.Token(r),
        "meta":         meta,
    })
	if err != nil {
		s.InternalServerError(w, r, err)
		return
	}
}
