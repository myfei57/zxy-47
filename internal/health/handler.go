package health

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	version string
}

func NewHandler(version string) *Handler {
	return &Handler{version: version}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Check(h.version))
}
