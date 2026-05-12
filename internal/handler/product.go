package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"orderfood/internal/model"
)

// ListProducts handles GET /product.
func ListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(model.Catalog())
}

// GetProduct handles GET /product/{productId}.
func GetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path: /product/<id>
	id := strings.TrimPrefix(r.URL.Path, "/product/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Product ID is required",
		})
		return
	}

	p := model.Lookup(id)
	if p == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "Product not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(p)
}

