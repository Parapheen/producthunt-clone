package handler

import (
	"net/http"
)

// Robots serves robots.txt with a link to the sitemap.
func (h *Handler) Robots(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")

    base := h.BaseURL

    // Allow all crawling, expose sitemap
    _, _ = w.Write([]byte("User-agent: *\nAllow: /\n\nSitemap: " + base + "/sitemap.xml\n"))
}


