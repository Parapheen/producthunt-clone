package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

// LaunchAwardsPage renders a page to manage awards for a specific launch by UUID
func (h *Handler) LaunchAwardsPage(w http.ResponseWriter, r *http.Request) {
	launchIDStr := r.PathValue("launchID")
	launchID, err := uuid.Parse(launchIDStr)
	if err != nil {
		http.Error(w, "invalid launch id", http.StatusBadRequest)
		return
	}

	l, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	awards, err := h.LaunchService.ListAwards(r.Context())
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	existing, _ := h.LaunchService.GetAwardsByLaunchIDs(r.Context(), []uuid.UUID{launchID})

	t, err := template.New("launch-awards").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string { return t.Format("2006-01-02") },
	}).ParseFiles(
		"views/admin/launch-awards.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	_ = t.ExecuteTemplate(w, "layout", map[string]any{
		"Launch":   l,
		"Awards":   awards,
		"Existing": existing[launchID],
		"token":    nosurf.Token(r),
	})
}

// LaunchAwardsPageByIndex renders awards page found by product slug and launch index
func (h *Handler) LaunchAwardsPageByIndex(w http.ResponseWriter, r *http.Request) {
	productSlug := r.PathValue("productSlug")
	indexStr := r.PathValue("index")
	idx, err := strconv.Atoi(indexStr)
	if err != nil || idx <= 0 {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	l, err := h.LaunchService.GetNthByProductOrderedByCreatedAt(r.Context(), p.ID, idx)
	if err != nil || l == nil {
		http.NotFound(w, r)
		return
	}

	awards, err := h.LaunchService.ListAwards(r.Context())
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	existing, _ := h.LaunchService.GetAwardsByLaunchIDs(r.Context(), []uuid.UUID{l.ID})

	t, err := template.New("launch-awards").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string { return t.Format("2006-01-02") },
	}).ParseFiles(
		"views/admin/launch-awards.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	_ = t.ExecuteTemplate(w, "layout", map[string]any{
		"Launch":   l,
		"Awards":   awards,
		"Existing": existing[l.ID],
		"token":    nosurf.Token(r),
	})
}

// AssignLaunchAward assigns an award from the dedicated page
func (h *Handler) AssignLaunchAward(w http.ResponseWriter, r *http.Request) {
	launchIDStr := r.PathValue("launchID")
	launchID, err := uuid.Parse(launchIDStr)
	if err != nil {
		http.Error(w, "invalid launch id", http.StatusBadRequest)
		return
	}
	awardCode := r.FormValue("award_code")
	period := r.FormValue("period_date") // YYYY-MM-DD
	if awardCode == "" || period == "" {
		http.Error(w, "award_code and period_date are required", http.StatusBadRequest)
		return
	}
	periodDate, err := time.Parse("2006-01-02", period)
	if err != nil {
		http.Error(w, "invalid period_date", http.StatusBadRequest)
		return
	}

	if err := h.LaunchService.AssignAwardToLaunch(r.Context(), launchID, awardCode, periodDate); err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	// On success, return toast for HTMX or redirect otherwise
	if r.Header.Get("HX-Request") == "true" {
		t, err := template.ParseFiles("views/partials/toast.html")
		if err != nil {
			h.InternalServerError(w, r, err)
			return
		}
		w.Header().Set("HX-Retarget", "body")
		w.Header().Set("HX-Reswap", "beforeend")
		_ = t.Execute(w, map[string]any{"Message": "Награда присвоена"})
		return
	}
	http.Redirect(w, r, "/admin/launches/"+launchID.String()+"/awards", http.StatusSeeOther)
}
