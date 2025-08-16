package handler

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	humantime "github.com/Parapheen/ph-clone/internal/pkg/human_time"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

// GetProductLaunchByIndex renders a launch by 1-based index within product launches ordered by created_at ASC (oldest first).
func (h *Handler) GetProductLaunchByIndex(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")
	indexStr := r.PathValue("index")
	idx, err := strconv.Atoi(indexStr)
	if err != nil || idx <= 0 {
		http.NotFound(w, r)
		return
	}

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	var l *launch.Launch
	l, err = h.LaunchService.GetNthByProductOrderedByCreatedAt(r.Context(), p.ID, idx)
	if err != nil || l == nil {
		http.NotFound(w, r)
		return
	}

	if !l.IsPublic() {
		http.NotFound(w, r)
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

	// Build makers view from product members
	memberRoles := make(map[uuid.UUID]string)
	for _, member := range p.Members {
		memberRoles[member.UserID] = member.Role.String()
	}

	memberIDs := make([]uuid.UUID, 0, len(memberRoles))
	for memberID := range memberRoles {
		memberIDs = append(memberIDs, memberID)
	}

	var makers []*MemberView
	if len(memberIDs) > 0 {
		users, err := h.UserService.GetByIDs(r.Context(), memberIDs)
		if err == nil {
			userMap := make(map[uuid.UUID]*user.User)
			for _, usr := range users {
				userMap[usr.ID] = usr
			}
			makers = make([]*MemberView, 0, len(memberIDs))
			for _, m := range p.Members {
				if usr, ok := userMap[m.UserID]; ok {
					makers = append(makers, &MemberView{ID: usr.ID, Name: usr.Name, Role: memberRoles[m.UserID], AvatarURL: usr.AvatarURL, Bio: usr.Bio})
				}
			}
		}
	}

	humanTime := humantime.HumanTime

	// Awards for this launch
	awardsByLaunch, _ := h.LaunchService.GetAwardsByLaunchIDs(r.Context(), []uuid.UUID{l.ID})

	t, err := template.New("launch.html").Funcs(template.FuncMap{
		"dict":           tmpl.Dict,
		"safeHTML":       func(s string) template.HTML { return template.HTML(s) },
		"humanTime":      humanTime,
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
		h.InternalServerError(w, r, err)
		return
	}

	canonical := h.BaseURL + "/products/" + p.Slug + "/launches/" + strconv.Itoa(idx)
	meta := map[string]any{
		"Title":       l.Name + " — запуск продукта " + p.Name,
		"Description": l.Tagline,
		"Canonical":   canonical,
		"OGType":      "article",
		"Image":       l.ImageURL,
	}

	err = t.ExecuteTemplate(w, "layout", map[string]any{
		"User":           u,
		"Product":        p,
		"Launch":         l,
		"Upvoted":        upvoted,
		"Makers":         makers,
		"token":          nosurf.Token(r),
		"meta":           meta,
		"AwardsByLaunch": awardsByLaunch,
	})
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}
