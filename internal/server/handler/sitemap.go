package handler

import (
	"encoding/xml"
	"net/http"
)

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []url    `xml:"url"`
}

type url struct {
	Loc string `xml:"loc"`
}

// Sitemap renders a minimal sitemap with home and product pages when possible.
func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	base := h.BaseURL

	urls := []url{
		{Loc: base + "/"},
	}

	// Attempt to list recent product slugs for discoverability
	// We intentionally avoid adding a repository method to keep change small:
	// fall back to no-op if direct query fails.
	type slugRow struct {
		Slug string `db:"slug"`
	}
	if h.DB != nil {
		rows := []slugRow{}
		// limit keeps sitemap light; adjust as needed
		_ = h.DB.Select(&rows, "SELECT slug FROM products ORDER BY created_at DESC LIMIT 500")
		for _, r := range rows {
			urls = append(urls, url{Loc: base + "/products/" + r.Slug})
		}

		// Include latest published launches under product slugs
		type launchRow struct {
			ProductSlug string `db:"product_slug"`
			LaunchSlug  string `db:"launch_slug"`
		}
		lrows := []launchRow{}
		_ = h.DB.Select(&lrows, `
            SELECT p.slug AS product_slug, l.slug AS launch_slug
            FROM launches l
            JOIN products p ON p.id = l.product_id
            WHERE l.state = 'published'
            ORDER BY COALESCE(l.launch_date, l.updated_at) DESC
            LIMIT 1000
        `)
		for _, r := range lrows {
			urls = append(urls, url{Loc: base + "/products/" + r.ProductSlug + "/launches/" + r.LaunchSlug})
		}
	}

	set := urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_ = xml.NewEncoder(w).Encode(set)
}
