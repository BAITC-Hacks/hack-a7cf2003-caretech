package httpapi

import (
	"encoding/json"
	"net/http"

	"hackalem-ekt-backend/internal/cart"
	"hackalem-ekt-backend/internal/catalog"
	"hackalem-ekt-backend/internal/chat"
)

type Handler struct {
	chatService *chat.Service
	catalog     catalog.Repository
	carts       cart.Store
}

func NewHandler(chatService *chat.Service, products catalog.Repository, carts cart.Store) http.Handler {
	api := &Handler{chatService: chatService, catalog: products, carts: carts}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.health)
	mux.HandleFunc("GET /api/products", api.products)
	mux.HandleFunc("POST /api/chat", api.chat)
	mux.HandleFunc("GET /api/cart", api.cart)
	mux.HandleFunc("POST /api/cart/confirm", api.confirmCart)
	return withJSON(withCORS(mux))
}

func (handler *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (handler *Handler) products(w http.ResponseWriter, request *http.Request) {
	writeJSON(w, http.StatusOK, handler.catalog.Search(request.URL.Query().Get("q")))
}

func (handler *Handler) chat(w http.ResponseWriter, request *http.Request) {
	var input chat.Request
	if !decodeJSON(w, request, &input) || input.SessionID == "" || input.Message == "" {
		return
	}
	writeJSON(w, http.StatusOK, handler.chatService.Handle(input))
}

func (handler *Handler) cart(w http.ResponseWriter, request *http.Request) {
	writeJSON(w, http.StatusOK, handler.carts.Get(request.URL.Query().Get("session_id")))
}

func (handler *Handler) confirmCart(w http.ResponseWriter, request *http.Request) {
	url, err := handler.chatService.Confirm(request.URL.Query().Get("session_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"cart_url": url})
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный JSON"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}