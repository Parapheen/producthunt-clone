package handler

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	humantime "github.com/Parapheen/ph-clone/internal/pkg/human_time"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	launches, err := h.LaunchService.GetPublishedByProduct(r.Context(), p.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}
	humanTime := humantime.HumanTime

	t, err := template.New("product.html").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"dict":           tmpl.Dict,
		"safeHTML":       func(s string) template.HTML { return template.HTML(s) },
		"humanTime":      humanTime,
		"formatDateTime": func(ts time.Time) string { return ts.Format("02.01.2006 15:04") },
	}).ParseFiles(
		"views/product/product.html",
		"views/product/partials/launches-tab.html",
		"views/product/partials/members-tab.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/launch-card.html",
		"views/partials/launch-state.html",
		"views/partials/launch-upvote.html",
		"views/partials/launch-comments.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	// Build upvoted map for current user, when service supports it
	var upvoted map[uuid.UUID]bool
	if u != nil {
		type upvoteAware interface {
			GetUpvotedMap(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error)
		}
		if svc, ok := any(h.LaunchService).(upvoteAware); ok {
			ids := make([]uuid.UUID, 0, len(launches))
			for _, l := range launches {
				ids = append(ids, l.ID)
			}
			up, _ := svc.GetUpvotedMap(r.Context(), u.ID, ids)
			upvoted = up
		}
	}

	// Awards by launch for this product
	launchIDs := make([]uuid.UUID, 0, len(launches))
	for _, l := range launches {
		launchIDs = append(launchIDs, l.ID)
	}
	awardsByLaunch, _ := h.LaunchService.GetAwardsByLaunchIDs(r.Context(), launchIDs)

	// Compute product awards as union of launch awards
	type awardView struct {
		Award struct{ Code, Name, Description, Icon string }
	}
	productAwards := make([]*awardView, 0)
	seen := map[string]struct{}{}
	for _, aws := range awardsByLaunch {
		for _, la := range aws {
			key := la.Award.Code + ":" + la.PeriodDate.Format("2006-01-02")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			productAwards = append(productAwards, &awardView{Award: struct{ Code, Name, Description, Icon string }{la.Award.Code, la.Award.Name, la.Award.Description, la.Award.Icon}})
		}
	}

	// SEO meta for Product
	canonical := h.BaseURL + "/products/" + p.Slug
	var image string
	if p.ImageURL != "" {
		image = p.ImageURL
	}
	meta := map[string]any{
		"Title":       p.Name + " — " + p.Tagline,
		"Description": p.Tagline,
		"Canonical":   canonical,
		"OGType":      "product",
		"Image":       image,
	}

	// Build index map (launch ID -> index) for created_at DESC ordering
	indexMap := make(map[uuid.UUID]int)
	type indexAware interface {
		GetIndexByProductAndLaunchID(ctx context.Context, productID, launchID uuid.UUID) (int, error)
	}
	if svc, ok := any(h.LaunchService).(indexAware); ok {
		for _, l := range launches {
			if idx, err := svc.GetIndexByProductAndLaunchID(r.Context(), p.ID, l.ID); err == nil {
				indexMap[l.ID] = idx
			}
		}
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":           u,
		"Product":        p,
		"Launches":       launches,
		"ActiveTab":      "launches",
		"token":          nosurf.Token(r),
		"UpvotedMap":     upvoted,
		"IndexMap":       indexMap,
		"meta":           meta,
		"AwardsByLaunch": awardsByLaunch,
		"ProductAwards":  productAwards,
	})
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

func (h *Handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productID")

	p, err := h.ProductService.GetByID(r.Context(), uuid.MustParse(productID))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/products/"+p.Slug, http.StatusFound)
}

type MemberView struct {
	ID        uuid.UUID
	Name      string
	Role      string
	AvatarURL string
	Bio       string
}

func (h *Handler) ProductMembers(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	memberRoles := make(map[uuid.UUID]string)
	for _, member := range p.Members {
		memberRoles[member.UserID] = member.Role.String()
	}

	memberIDs := make([]uuid.UUID, 0, len(memberRoles))
	for memberID := range memberRoles {
		memberIDs = append(memberIDs, memberID)
	}

	users, err := h.UserService.GetByIDs(r.Context(), memberIDs)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting users", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	userMap := make(map[uuid.UUID]*user.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	membersView := make([]*MemberView, 0, len(p.Members))
	for _, member := range p.Members {
		user, ok := userMap[member.UserID]
		if !ok {
			continue
		}
		membersView = append(membersView, &MemberView{
			ID:        user.ID,
			Name:      user.Name,
			Role:      memberRoles[member.UserID],
			AvatarURL: user.AvatarURL,
			Bio:       user.Bio,
		})
	}

	if r.Header.Get("HX-Request") == "true" {
		t, err := template.New("product-members.html").
			Funcs(template.FuncMap{
				"dict": tmpl.Dict,
			}).
			ParseFiles(
				"views/product/partials/members-tab.html",
			)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "error parsing template", slog.Any("error", err))
			h.InternalServerError(w, r, err)
			return
		}

		err = t.ExecuteTemplate(w, "members-tab", map[string]interface{}{
			"User":    u,
			"Members": membersView,
			"token":   nosurf.Token(r),
		})
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "error executing template", slog.Any("error", err))
			h.InternalServerError(w, r, err)
			return
		}
		return
	}

	launches, err := h.LaunchService.GetByProduct(r.Context(), p.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	humanTime := humantime.HumanTime

	// Awards for these launches
	launchIDs := make([]uuid.UUID, 0, len(launches))
	for _, l := range launches {
		launchIDs = append(launchIDs, l.ID)
	}
	awardsByLaunch, _ := h.LaunchService.GetAwardsByLaunchIDs(r.Context(), launchIDs)

	t, err := template.New("product.html").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"dict":           tmpl.Dict,
		"humanTime":      humanTime,
		"formatDateTime": func(ts time.Time) string { return ts.Format("02.01.2006 15:04") },
	}).ParseFiles(
		"views/product/product.html",
		"views/product/partials/launches-tab.html",
		"views/product/partials/members-tab.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/launch-card.html",
		"views/partials/launch-upvote.html",
	)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error parsing template", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	// SEO for members tab uses same product canonical
	canonical := h.BaseURL + "/products/" + p.Slug
	var image string
	if p.ImageURL != "" {
		image = p.ImageURL
	}
	meta := map[string]any{
		"Title":       p.Name + " — команда продукта",
		"Description": p.Tagline,
		"Canonical":   canonical,
		"OGType":      "product",
		"Image":       image,
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":           u,
		"Members":        membersView,
		"Product":        p,
		"Launches":       launches,
		"ActiveTab":      "members",
		"token":          nosurf.Token(r),
		"meta":           meta,
		"AwardsByLaunch": awardsByLaunch,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error executing template", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}
}

func (s *Handler) ProductLaunches(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := s.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		s.InternalServerError(w, r, err)
		return
	}

	launches, err := s.LaunchService.GetPublishedByProduct(r.Context(), p.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		s.InternalServerError(w, r, err)
		return
	}

	// Build upvoted map for current user, when service supports it
	var upvoted map[uuid.UUID]bool
	if user != nil {
		type upvoteAware interface {
			GetUpvotedMap(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error)
		}
		if svc, ok := any(s.LaunchService).(upvoteAware); ok {
			ids := make([]uuid.UUID, 0, len(launches))
			for _, l := range launches {
				ids = append(ids, l.ID)
			}
			up, _ := svc.GetUpvotedMap(r.Context(), user.ID, ids)
			upvoted = up
		}
	}

	humanTime := humantime.HumanTime

	if r.Header.Get("HX-Request") == "true" {
		t, err := template.New("product-launches.html").
			Funcs(template.FuncMap{
				"dict":           tmpl.Dict,
				"safeHTML":       func(s string) template.HTML { return template.HTML(s) },
				"humanTime":      humanTime,
				"formatDateTime": func(ts time.Time) string { return ts.Format("02.01.2006 15:04") },
			}).
			ParseFiles(
				"views/product/partials/launches-tab.html",
				"views/partials/launch-state.html",
				"views/partials/launch-card.html",
				"views/partials/launch-upvote.html",
				"views/partials/launch-comments.html",
			)
		if err != nil {
			s.InternalServerError(w, r, err)
			return
		}

		// Awards for these launches
		launchIDs := make([]uuid.UUID, 0, len(launches))
		for _, l := range launches {
			launchIDs = append(launchIDs, l.ID)
		}
		awardsByLaunch, _ := s.LaunchService.GetAwardsByLaunchIDs(r.Context(), launchIDs)

		err = t.ExecuteTemplate(w, "launches-tab", map[string]interface{}{
			"User":           user,
			"Launches":       launches,
			"UpvotedMap":     upvoted,
			"token":          nosurf.Token(r),
			"AwardsByLaunch": awardsByLaunch,
		})
		if err != nil {
			s.InternalServerError(w, r, err)
			return
		}
		return
	}

	t, err := template.New("product-launches.html").
		Funcs(template.FuncMap{
			"dict":           tmpl.Dict,
			"safeHTML":       func(s string) template.HTML { return template.HTML(s) },
			"humanTime":      humanTime,
			"formatDateTime": func(ts time.Time) string { return ts.Format("02.01.2006 15:04") },
		}).
		ParseFiles(
			"views/product/product.html",
			"views/product/partials/launches-tab.html",
			"views/product/partials/members-tab.html",
			"views/layout/layout.html",
			"views/layout/header.html",
			"views/layout/footer.html",
			"views/layout/head.html",
			"views/partials/launch-card.html",
			"views/partials/launch-state.html",
			"views/partials/launch-upvote.html",
		)
	if err != nil {
		s.InternalServerError(w, r, err)
		return
	}

	// Awards for these launches
	launchIDs := make([]uuid.UUID, 0, len(launches))
	for _, l := range launches {
		launchIDs = append(launchIDs, l.ID)
	}
	awardsByLaunch, _ := s.LaunchService.GetAwardsByLaunchIDs(r.Context(), launchIDs)

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":           user,
		"Launches":       launches,
		"Product":        p,
		"ActiveTab":      "launches",
		"token":          nosurf.Token(r),
		"UpvotedMap":     upvoted,
		"AwardsByLaunch": awardsByLaunch,
	})
	if err != nil {
		s.InternalServerError(w, r, err)
		return
	}
}
