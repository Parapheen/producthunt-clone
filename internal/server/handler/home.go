package handler

import (
	"html/template"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

func (s *Handler) Home(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	t, err := template.ParseFiles("views/index.html", "views/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User": user,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
