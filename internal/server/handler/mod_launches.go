package handler

import (
	"html/template"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/justinas/nosurf"
)

func (h *Handler) ModLaunches(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	launches, err := h.LaunchService.GetByState(r.Context(), []launch.State{launch.Review, launch.Declined})

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	t, err := template.New("product-launches.html").
		Funcs(template.FuncMap{
			"dict": tmpl.Dict,
		}).
		ParseFiles(
			"views/moderation-launches.html",
			"views/partials/launch-state.html",
			"views/layout/layout.html",
			"views/layout/header.html",
			"views/layout/footer.html",
			"views/layout/head.html",
			"views/partials/launch-moderation-card.html",
		)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":     user,
		"Launches": launches,
		"token":    nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
