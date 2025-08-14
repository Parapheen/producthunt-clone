package handler

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	neturl "net/url"
	"sort"
	"strings"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/validation"
	"github.com/justinas/nosurf"
	"golang.org/x/net/html"
)

func (h *Handler) NewProductForm(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	t, err := template.ParseFiles(
		"views/new-product.html",
		"views/partials/select-categories.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":  u,
		"token": nosurf.Token(r),
	})
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

func (h *Handler) NewProduct(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	if u == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errors := make([]string, 0)

	name := r.FormValue("name")
	productURL := r.FormValue("url")
	tagline := r.FormValue("tagline")

	// Validate input fields early
	v := validation.NewValidator()
	if verr := v.ValidateMultiple(
		v.ValidateString(name, "name", 1, product.ProductNameMaxLength, true),
		v.ValidateURL(productURL, "url", true),
		v.ValidateString(tagline, "tagline", 0, 140, false),
	); verr != nil {
		switch ve := verr.(type) {
		case validation.ValidationErrors:
			for _, e := range ve {
				errors = append(errors, e.Error())
			}
		default:
			errors = append(errors, verr.Error())
		}
	}

	nameExists, err := h.ProductService.NameExists(r.Context(), name)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error checking if product name exists", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	urlExists, err := h.ProductService.URLExists(r.Context(), productURL)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error checking if product url exists", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	if nameExists {
		errors = append(errors, "Продукт с таким названием уже существует")
	}

	if urlExists {
		errors = append(errors, "Продукт с таким URL уже существует")
	}

	if len(errors) > 0 {
		h.renderErrors(w, r, errors)
		return
	}

	h.Logger.InfoContext(r.Context(), "creating product", slog.Any("name", name), slog.Any("url", productURL))

	categories := make([]*product.Category, 0)
	for _, category := range r.PostForm["categories"] {
		c, err := h.ProductService.GetCategoryBySlug(r.Context(), category)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "error getting category", slog.Any("error", err))
			errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
			continue
		}
		categories = append(categories, c)
	}

	p := product.NewProduct(name, productURL, tagline, categories, u.ID)

	err = h.ProductService.Create(
		r.Context(),
		p,
	)

	switch err {
	case nil:
        // Fire-and-forget favicon fetch and set as product image (try best-quality sources first)
		prodID := p.ID
		rawURL := productURL
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
			defer cancel()

            parsedURL, perr := neturl.Parse(rawURL)
            if perr != nil || parsedURL.Host == "" {
				return
			}

			client := &http.Client{Timeout: 5 * time.Second}

            // Helper: decide filename from content type
            detectFilename := func(contentType string) string {
                ct := strings.ToLower(contentType)
                switch {
                case strings.Contains(ct, "svg"):
                    return "favicon.svg"
                case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
                    return "favicon.jpg"
                case strings.Contains(ct, "png"):
                    return "favicon.png"
                case strings.Contains(ct, "gif"):
                    return "favicon.gif"
                default:
                    return "favicon.ico"
                }
            }

            // Helper: fetch URL and set as product image
            fetchAndSet := func(iconURL string) bool {
                req, err := http.NewRequestWithContext(ctx, http.MethodGet, iconURL, nil)
                if err != nil {
                    return false
                }
                req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ph-clone/1.0)")
                req.Header.Set("Accept", "image/*,application/octet-stream;q=0.8,*/*;q=0.5")
                resp, err := client.Do(req)
                if err != nil {
                    return false
                }
                defer resp.Body.Close()
                if resp.StatusCode != http.StatusOK {
                    return false
                }
                filename := detectFilename(resp.Header.Get("Content-Type"))
                if _, upErr := h.ProductService.UpdateImage(ctx, prodID, filename, resp.Body); upErr == nil {
                    return true
                }
                return false
            }

            // Helper: resolve relative URLs
            resolve := func(base *neturl.URL, href string) string {
                if href == "" {
                    return ""
                }
                u, err := neturl.Parse(href)
                if err != nil {
                    return ""
                }
                if u.IsAbs() {
                    return u.String()
                }
                return base.ResolveReference(u).String()
            }

            // Discover icons from HTML link tags, prefer largest sizes
            type iconCandidate struct {
                url  string
                size int // max dimension; "any" treated as large (512)
            }

            discoverIcons := func(page string) []iconCandidate {
                req, err := http.NewRequestWithContext(ctx, http.MethodGet, page, nil)
                if err != nil {
                    return nil
                }
                req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ph-clone/1.0)")
                resp, err := client.Do(req)
                if err != nil {
                    return nil
                }
                defer resp.Body.Close()
                if resp.StatusCode != http.StatusOK || !strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
                    return nil
                }
                doc, err := html.Parse(resp.Body)
                if err != nil {
                    return nil
                }
                baseURL := req.URL
                candidates := make([]iconCandidate, 0, 8)
                var walk func(*html.Node)
                walk = func(n *html.Node) {
                    if n.Type == html.ElementNode && n.Data == "link" {
                        var rel, href, sizes string
                        for _, a := range n.Attr {
                            switch strings.ToLower(a.Key) {
                            case "rel":
                                rel = strings.ToLower(a.Val)
                            case "href":
                                href = a.Val
                            case "sizes":
                                sizes = strings.ToLower(a.Val)
                            }
                        }
                        if rel != "" && (strings.Contains(rel, "icon") || strings.Contains(rel, "apple-touch-icon") || strings.Contains(rel, "mask-icon")) {
                            abs := resolve(baseURL, href)
                            if abs == "" {
                                // skip
                            } else {
                                size := 0
                                if sizes == "any" {
                                    size = 512
                                } else if sizes != "" {
                                    // parse patterns like "32x32", or "180x180 152x152"
                                    parts := strings.Fields(sizes)
                                    for _, pz := range parts {
                                        if xy := strings.Split(pz, "x"); len(xy) == 2 {
                                            // best-effort parse
                                            if a, b := strings.TrimSpace(xy[0]), strings.TrimSpace(xy[1]); a != "" && b != "" {
                                                // convert to int
                                                var ai, bi int
                                                for i := 0; i < len(a); i++ {
                                                    if a[i] < '0' || a[i] > '9' {
                                                        ai = 0
                                                        break
                                                    }
                                                    ai = ai*10 + int(a[i]-'0')
                                                }
                                                for i := 0; i < len(b); i++ {
                                                    if b[i] < '0' || b[i] > '9' {
                                                        bi = 0
                                                        break
                                                    }
                                                    bi = bi*10 + int(b[i]-'0')
                                                }
                                                if ai > size {
                                                    if ai >= bi {
                                                        size = ai
                                                    } else {
                                                        size = bi
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                                if size == 0 {
                                    size = 64 // unknown; assume small-ish
                                }
                                candidates = append(candidates, iconCandidate{url: abs, size: size})
                            }
                        }
                    }
                    for c := n.FirstChild; c != nil; c = c.NextSibling {
                        walk(c)
                    }
                }
                walk(doc)
                sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].size > candidates[j].size })
                return candidates
            }

            // 1) Try icons discovered from the product page itself
            for _, c := range discoverIcons(rawURL) {
                if fetchAndSet(c.url) {
							return
						}
					}

            // 2) Try icons discovered from site root
            siteRoot := parsedURL.Scheme + "://" + parsedURL.Host + "/"
            for _, c := range discoverIcons(siteRoot) {
                if fetchAndSet(c.url) {
                    return
                }
            }

            // 3) Try common high-res paths
            commonPaths := []string{
                "/android-chrome-512x512.png",
                "/apple-touch-icon.png",
                "/favicon-32x32.png",
                "/favicon-16x16.png",
                "/favicon.png",
                "/favicon.ico",
            }
            for _, path := range commonPaths {
                candidate := parsedURL.Scheme + "://" + parsedURL.Host + path
                if fetchAndSet(candidate) {
                    return
                }
            }

            // 4) Fallback to Google S2 favicon (request larger size)
            fallback := "https://www.google.com/s2/favicons?sz=256&domain_url=" + neturl.QueryEscape(siteRoot)
            _ = fetchAndSet(fallback)
		}()

		w.Header().Add("HX-Redirect", "/products/"+p.ID.String()+"/new-launch")
		return
	case product.ErrProductNameTooLong:
		errors = append(errors, "Название продукта слишком длинное")
	case product.ErrProductURLTooLong:
		errors = append(errors, "URL продукта слишком длинный")
	case product.ErrInvalidURLScheme, product.ErrInvalidURL:
		errors = append(errors, "Невалидный URL")
	case product.ErrProductNameEmpty:
		errors = append(errors, "Название продукта не может быть пустым")
	case product.ErrProductURLEmpty:
		errors = append(errors, "URL продукта не может быть пустым")
	case product.ErrCategoryNotFound:
		errors = append(errors, "Категория не найдена")
	case product.ErrNoCategories:
		errors = append(errors, "Необходимо добавить хотя бы одну категорию")
	case product.ErrTooManyCategories:
		errors = append(errors, "Не более 3 категорий")
	default:
		h.Logger.ErrorContext(r.Context(), "error creating product", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	if len(errors) > 0 {
		h.renderErrors(w, r, errors)
		return
	}
}

func (h *Handler) renderErrors(w http.ResponseWriter, r *http.Request, errors []string) {
	t, err := template.ParseFiles("views/partials/errors.html")
	if err != nil {
		h.Logger.Error("failed to parse errors template", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"Errors": errors,
	})
	if err != nil {
		h.Logger.Error("failed to execute errors template", slog.Any("error", err))
		h.InternalServerError(w, r, err)
	}
}