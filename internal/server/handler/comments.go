package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	humantime "github.com/Parapheen/ph-clone/internal/pkg/human_time"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

// GetLaunchComments renders the HTMX partial with comments and editor
func (h *Handler) GetLaunchComments(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

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

	// Fetch comments
	roots, replies, err := h.LaunchService.GetCommentsTree(r.Context(), l.ID)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	// Fetch minimal user data for authors
	authorIDSet := map[uuid.UUID]struct{}{}
	for _, c := range roots {
		authorIDSet[c.AuthorID] = struct{}{}
		for _, rc := range replies[c.ID] {
			authorIDSet[rc.AuthorID] = struct{}{}
		}
	}
	authorIDs := make([]uuid.UUID, 0, len(authorIDSet))
	for id := range authorIDSet {
		authorIDs = append(authorIDs, id)
	}
	authors, _ := h.UserService.GetByIDs(r.Context(), authorIDs)
	authorMap := make(map[uuid.UUID]*user.User)
	for _, au := range authors {
		authorMap[au.ID] = au
	}

	// Build member role map for this product to annotate authors
	memberRoleByUser := make(map[uuid.UUID]string)
	if p, err := h.ProductService.GetByID(r.Context(), l.ProductID); err == nil {
		for _, m := range p.Members {
			memberRoleByUser[m.UserID] = m.Role.String()
		}
	}

	// Determine owner permission for pin button
	canPin := false
	if u != nil {
		if p, err := h.ProductService.GetByID(r.Context(), l.ProductID); err == nil {
			if p.IsOwner(u.ID) {
				canPin = true
			}
		}
	}

	// Helper to format human-readable relative time in Russian.
	humanTime := humantime.HumanTime

	t, err := template.New("launch-comments").Funcs(template.FuncMap{
		"safeHTML":       func(s string) template.HTML { return template.HTML(s) },
		"dict":           tmpl.Dict,
		"humanTime":      humanTime,
		"formatDateTime": func(ts time.Time) string { return ts.Format("02.01.2006 15:04") },
	}).ParseFiles(
		"views/partials/launch-comments.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	_ = t.ExecuteTemplate(w, "launch-comments", map[string]any{
		"Launch":      l,
		"Roots":       roots,
		"Replies":     replies,
		"Authors":     authorMap,
		"MemberRoles": memberRoleByUser,
		"token":       nosurf.Token(r),
		"CurrentUser": u,
		"CanPin":      canPin,
	})
}

// PostLaunchComment handles creation of a top-level comment
func (h *Handler) PostLaunchComment(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		unauthorizedAuthModal(w)
		return
	}
	launchID, err := uuid.Parse(r.PathValue("launchID"))
	if err != nil {
		http.Error(w, "invalid launch id", http.StatusBadRequest)
		return
	}
	content := r.FormValue("content_html")
	tag := r.FormValue("tag")
	c := launch.NewComment(launchID, u.ID, content)
	switch tag {
	case string(launch.CommentTagIdea), string(launch.CommentTagQuestion), string(launch.CommentTagLike):
		c.Tag = launch.CommentTag(tag)
	default:
		http.Error(w, "invalid tag", http.StatusBadRequest)
		return
	}
	if err := h.LaunchService.CreateComment(r.Context(), c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Email notification to product owners about new comment
	if h.Mailer != nil {
		if l, err := h.LaunchService.GetByID(r.Context(), launchID); err == nil {
			if p, pErr := h.ProductService.GetByID(r.Context(), l.ProductID); pErr == nil {
				recipients := make([]uuid.UUID, 0)
				for _, m := range p.Members {
					if m.UserID == u.ID {
						continue
					}
					recipients = append(recipients, m.UserID)
				}
				if len(recipients) > 0 {
					if members, uErr := h.UserService.GetByIDs(r.Context(), recipients); uErr == nil {
						index, _ := h.LaunchService.GetIndexByProductAndLaunchID(r.Context(), p.ID, l.ID)
						launchURL := h.BaseURL + "/products/" + p.Slug + "/launches/" + strconv.Itoa(index)
						data := map[string]any{
							"ProductName": p.Name,
							"LaunchName":  l.Name,
							"LaunchURL":   launchURL,
							"AuthorName":  u.Name,
							"ContentHTML": c.ContentHTML,
						}
						for _, member := range members {
							_ = h.Mailer.Send(r.Context(), member.Email, "launch_new_comment.html", data)
						}
					}
				}
			}
		}
	}
	h.GetLaunchComments(w, r)
}

// ReplyLaunchComment handles creation of a reply to a root comment
func (h *Handler) ReplyLaunchComment(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		unauthorizedAuthModal(w)
		return
	}
	launchID, err := uuid.Parse(r.PathValue("launchID"))
	if err != nil {
		http.Error(w, "invalid launch id", http.StatusBadRequest)
		return
	}
	parentID, err := uuid.Parse(r.PathValue("commentID"))
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}
	// Only allow replying to top-level comments: we rely on UI to provide only top-level IDs.
	content := r.FormValue("content_html")
	c := launch.NewComment(launchID, u.ID, content)
	c.ParentID = &parentID
	if err := h.LaunchService.CreateComment(r.Context(), c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	productOwnerID := uuid.Nil

	// Email notification to product members about new reply
	if h.Mailer != nil {
		if l, err := h.LaunchService.GetByID(r.Context(), launchID); err == nil {
			if p, pErr := h.ProductService.GetByID(r.Context(), l.ProductID); pErr == nil {
				recipients := make([]uuid.UUID, 0)
				for _, m := range p.Members {
					if m.Role == product.Owner {
						productOwnerID = m.UserID
					}
					if m.UserID == u.ID {
						continue
					}
					recipients = append(recipients, m.UserID)
				}
				if len(recipients) > 0 {
					if members, uErr := h.UserService.GetByIDs(r.Context(), recipients); uErr == nil {
						index, _ := h.LaunchService.GetIndexByProductAndLaunchID(r.Context(), p.ID, l.ID)
						launchURL := h.BaseURL + "/products/" + p.Slug + "/launches/" + strconv.Itoa(index)
						data := map[string]any{
							"ProductName": p.Name,
							"LaunchName":  l.Name,
							"LaunchURL":   launchURL,
							"AuthorName":  u.Name,
							"ContentHTML": c.ContentHTML,
						}
						for _, member := range members {
							_ = h.Mailer.Send(r.Context(), member.Email, "launch_new_comment.html", data)
						}
					}
				}
			}
		}
	}

	// Email notification to the parent comment author about received reply
	if h.Mailer != nil {
		if l, err := h.LaunchService.GetByID(r.Context(), launchID); err == nil {
			// Get root comments and find the parent author
			if roots, _, err := h.LaunchService.GetCommentsTree(r.Context(), l.ID); err == nil {
				var parentAuthorID uuid.UUID
				for _, root := range roots {
					// Skip if the parent comment is by the product owner
					// since he will receive notification about new comment
					if root.ID == parentID && root.AuthorID != productOwnerID {
						parentAuthorID = root.AuthorID
						break
					}
				}
				if parentAuthorID != uuid.Nil && parentAuthorID != u.ID {
					if parentUser, err := h.UserService.GetByID(r.Context(), parentAuthorID); err == nil {
						if p, pErr := h.ProductService.GetByID(r.Context(), l.ProductID); pErr == nil {
							index, _ := h.LaunchService.GetIndexByProductAndLaunchID(r.Context(), p.ID, l.ID)
							launchURL := h.BaseURL + "/products/" + p.Slug + "/launches/" + strconv.Itoa(index) + "#comments-" + l.ID.String()
							data := map[string]any{
								"ProductName": p.Name,
								"LaunchName":  l.Name,
								"LaunchURL":   launchURL,
								"AuthorName":  u.Name,
								"ContentHTML": c.ContentHTML,
							}
							_ = h.Mailer.Send(r.Context(), parentUser.Email, "launch_reply_comment.html", data)
						}
					}
				}
			}
		}
	}
	h.GetLaunchComments(w, r)
}

// ToggleCommentUpvote toggles comment upvote and returns the single comment HTML
// Comment upvote feature removed (UI and routes). Handler intentionally omitted.

// Pin or unpin a comment. Allowed for product owner only.
func (h *Handler) TogglePinComment(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	launchID, err := uuid.Parse(r.PathValue("launchID"))
	if err != nil {
		http.Error(w, "invalid launch id", http.StatusBadRequest)
		return
	}
	// Check ownership via product
	l, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	p, err := h.ProductService.GetByID(r.Context(), l.ProductID)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	if !p.IsOwner(u.ID) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	commentID, err := uuid.Parse(r.PathValue("commentID"))
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}
	pinned := r.FormValue("pinned") == "true"
	if err := h.LaunchService.PinComment(r.Context(), commentID, pinned); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	// Re-render the full comments block to reflect pin ordering
	h.GetLaunchComments(w, r)
}

func unauthorizedAuthModal(w http.ResponseWriter) {
	t, err := template.ParseFiles("views/partials/auth-modal.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Retarget", "body")
	w.Header().Set("HX-Reswap", "beforeend")
	_ = t.Execute(w, nil)
}
