package handler

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/justinas/nosurf"

	"github.com/Parapheen/ph-clone/internal/domain/blog"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
)

// AdminBlogsIndex shows a simple list of published posts for now
func (h *Handler) AdminBlogsIndex(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	posts, err := h.BlogService.ListPublished(r.Context(), 100, 0)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	t, err := template.New("admin-blogs").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/admin/blogs.html",
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
		"Title":     "Админ — блог",
		"Canonical": h.BaseURL + "/admin/blogs",
	}

	if err := t.ExecuteTemplate(w, "layout", map[string]any{
		"User":  u,
		"Posts": posts,
		"token": nosurf.Token(r),
		"meta":  meta,
	}); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

// AdminBlogsNew renders a form to create a blog post
func (h *Handler) AdminBlogsNew(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	t, err := template.New("admin-blogs-new").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/admin/blogs-new.html",
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
		"Title":     "Новый пост — блог",
		"Canonical": h.BaseURL + "/admin/blogs/new",
	}

	if err := t.ExecuteTemplate(w, "layout", map[string]any{
		"User":  u,
		"token": nosurf.Token(r),
		"meta":  meta,
	}); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

// AdminBlogsCreate handles form submission to create a blog post
func (h *Handler) AdminBlogsCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	excerpt := strings.TrimSpace(r.FormValue("excerpt"))
	content := strings.TrimSpace(r.FormValue("content"))
	published := r.FormValue("published") == "on"

	p := blog.NewPost(title, excerpt, content, published)
	if err := h.BlogService.Create(r.Context(), p); err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/blogs", http.StatusFound)
}

// AdminBlogsEdit renders a form to edit an existing blog post
func (h *Handler) AdminBlogsEdit(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	slug := r.PathValue("slug")
	p, err := h.BlogService.GetRawBySlug(r.Context(), slug)
	if err != nil || p == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	t, err := template.New("admin-blogs-edit").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/admin/blogs-edit.html",
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
		"Title":     "Редактировать пост — блог",
		"Canonical": h.BaseURL + "/admin/blogs/" + p.Slug + "/edit",
	}

	if err := t.ExecuteTemplate(w, "layout", map[string]any{
		"User":  u,
		"Post":  p,
		"token": nosurf.Token(r),
		"meta":  meta,
	}); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

// AdminBlogsUpdate handles form submission to update an existing blog post
func (h *Handler) AdminBlogsUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	slug := r.PathValue("slug")
	p, err := h.BlogService.GetRawBySlug(r.Context(), slug)
	if err != nil || p == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	excerpt := strings.TrimSpace(r.FormValue("excerpt"))
	content := strings.TrimSpace(r.FormValue("content"))
	published := r.FormValue("published") == "on"

	p.Title = title
	p.Slug = slugify.Slugify(title)
	p.Excerpt = excerpt
	p.ContentMD = content
	p.Published = published

	if err := h.BlogService.Update(r.Context(), p); err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/blogs", http.StatusFound)
}
