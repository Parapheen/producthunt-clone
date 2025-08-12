package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
)

type uploadResponse struct {
    URL string `json:"url"`
}

// UploadUserAvatar handles avatar uploads for a user.
func (h *Handler) UploadUserAvatar(w http.ResponseWriter, r *http.Request) {
    u := user.GetUserFromContext(r.Context())
    if u == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    userIDParam := r.PathValue("userID")
    userID, err := uuid.Parse(userIDParam)
    if err != nil || userID != u.ID {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    file, header, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "file is required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Forward to application service to store and persist URL
    url, err := h.UserService.UpdateAvatar(r.Context(), u.ID, header.Filename, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(uploadResponse{URL: url})
}

// UploadProductImage handles product image uploads (owner only).
func (h *Handler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
    u := user.GetUserFromContext(r.Context())
    if u == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    productIDParam := r.PathValue("productID")
    productID, err := uuid.Parse(productIDParam)
    if err != nil {
        http.Error(w, "invalid product id", http.StatusBadRequest)
        return
    }

    p, err := h.ProductService.GetByID(r.Context(), productID)
    if err != nil || p == nil {
        http.Error(w, "product not found", http.StatusNotFound)
        return
    }
    if !p.IsOwner(u.ID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    file, header, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "file is required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    url, err := h.ProductService.UpdateImage(r.Context(), p.ID, header.Filename, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(uploadResponse{URL: url})
}

// UploadLaunchMedia handles launch media uploads (owner only).
func (h *Handler) UploadLaunchMedia(w http.ResponseWriter, r *http.Request) {
    u := user.GetUserFromContext(r.Context())
    if u == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    launchIDParam := r.PathValue("launchID")
    launchID, err := uuid.Parse(launchIDParam)
    if err != nil {
        http.Error(w, "invalid launch id", http.StatusBadRequest)
        return
    }

    l, err := h.LaunchService.GetByID(r.Context(), launchID)
    if err != nil || l == nil {
        http.Error(w, "launch not found", http.StatusNotFound)
        return
    }

    // Verify ownership via product
    p, err := h.ProductService.GetByID(r.Context(), l.ProductID)
    if err != nil || p == nil || !p.IsOwner(u.ID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    file, header, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "file is required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Read into a Tee if needed; storage handles streaming
    var reader io.Reader = file
    url, err := h.LaunchService.AddMedia(r.Context(), l, header.Filename, reader)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(uploadResponse{URL: url})
}


