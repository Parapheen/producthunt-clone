package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	launches, err := h.LaunchService.GetPublishedByProduct(r.Context(), p.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.New("product.html").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"dict": tmpl.Dict,
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
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":      u,
		"Product":   p,
		"Launches":  launches,
		"ActiveTab": "launches",
		"token":     nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productID")

	p, err := h.ProductService.GetByID(r.Context(), uuid.MustParse(productID))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products/"+p.Slug, http.StatusFound)
}

type MemberView struct {
	Name string
	Role string
}

func (h *Handler) ProductMembers(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	members, err := h.ProductService.GetMembers(r.Context(), p.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting members", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	memberRoles := make(map[uuid.UUID]string)
	for _, member := range members {
		memberRoles[member.UserID] = member.Role.String()
	}

	memberIDs := make([]uuid.UUID, 0, len(memberRoles))
	for memberID := range memberRoles {
		memberIDs = append(memberIDs, memberID)
	}

	users, err := h.UserService.GetByIDs(r.Context(), memberIDs)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting users", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userMap := make(map[uuid.UUID]*user.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	membersView := make([]*MemberView, 0, len(members))
	for _, member := range members {
		user, ok := userMap[member.UserID]
		if !ok {
			continue
		}
		membersView = append(membersView, &MemberView{
			Name: user.Name,
			Role: memberRoles[member.UserID],
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = t.ExecuteTemplate(w, "members-tab", map[string]interface{}{
			"User":    u,
			"Members": membersView,
			"token":   nosurf.Token(r),
		})
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "error executing template", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	launches, err := h.LaunchService.GetByProduct(r.Context(), p.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.New("product.html").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/product/product.html",
		"views/product/partials/launches-tab.html",
		"views/product/partials/members-tab.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/launch-card.html",
	)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error parsing template", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":      u,
		"Members":   membersView,
		"Product":   p,
		"Launches":  launches,
		"ActiveTab": "members",
		"token":     nosurf.Token(r),
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error executing template", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Handler) ProductLaunches(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := s.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !p.IsOwner(user.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	launches, err := s.LaunchService.GetPublishedByProduct(r.Context(), p.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		t, err := template.New("product-launches.html").
			Funcs(template.FuncMap{
				"dict": tmpl.Dict,
			}).
			ParseFiles(
				"views/product/partials/launches-tab.html",
				"views/partials/launch-state.html",
				"views/partials/launch-card.html",
			)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = t.ExecuteTemplate(w, "launches-tab", map[string]interface{}{
			"User":     user,
			"Launches": launches,
			"token":    nosurf.Token(r),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	t, err := template.New("product-launches.html").
		Funcs(template.FuncMap{
			"dict": tmpl.Dict,
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
		)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":      user,
		"Launches":  launches,
		"Product":   p,
		"ActiveTab": "launches",
		"token":     nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
