package handler

import (
	"html/template"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
)


func (s *Handler) Rules(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

    t, err := template.New("rules").Funcs(template.FuncMap{
        "dict": tmpl.Dict,
    }).ParseFiles(
        "views/rules.html",
        "views/layout/layout.html",
        "views/layout/header.html",
        "views/layout/footer.html",
        "views/layout/head.html",
    )
    if err != nil {
        s.InternalServerError(w, r, err)
        return
    }

    // SEO meta for Rules
    baseURL := s.BaseURL
    meta := map[string]any{
        "Title":       "Пользовательское соглашение — justlaunch",
        "Description": "Пользовательское соглашение для justlaunch",
        "Canonical":   baseURL + "/",
        "OGType":      "website",
    }

    err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
        "User":         user,
        "meta":         meta,
    })
    if err != nil {
        s.InternalServerError(w, r, err)
        return
    }
}


func (s *Handler) Policy(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

    t, err := template.New("policy").Funcs(template.FuncMap{
        "dict": tmpl.Dict,
    }).ParseFiles(
        "views/policy.html",
        "views/layout/layout.html",
        "views/layout/header.html",
        "views/layout/footer.html",
        "views/layout/head.html",
    )
    if err != nil {
        s.InternalServerError(w, r, err)
        return
    }

    // SEO meta for Policy
    baseURL := s.BaseURL
    meta := map[string]any{
        "Title":       "Политика конфиденциальности — justlaunch",
        "Description": "Политика конфиденциальности для justlaunch",
        "Canonical":   baseURL + "/",
        "OGType":      "website",
    }

    err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
        "User":         user,
        "meta":         meta,
    })
    if err != nil {
        s.InternalServerError(w, r, err)
        return
    }
}