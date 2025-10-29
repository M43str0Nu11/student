package api

import (
	"encoding/json"
	"net/http"

	"orders-demo/internal/cache"

	"github.com/gorilla/mux"
)

type Handler struct {
	cache *cache.Cache
}

func NewHandler(c *cache.Cache) *Handler {
	return &Handler{cache: c}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	order, ok := h.cache.Get(id)
	if !ok {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *Handler) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/orders/{id}", h.GetOrder).Methods("GET")
	return r
}
