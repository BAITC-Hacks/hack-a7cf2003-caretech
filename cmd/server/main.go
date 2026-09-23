package main

import (
	"log"
	"net/http"

	"hackalem-ekt-backend/internal/cart"
	"hackalem-ekt-backend/internal/catalog"
	"hackalem-ekt-backend/internal/chat"
	httpapi "hackalem-ekt-backend/internal/httpapi"
)

func main() {
	catalogStore := catalog.NewMemoryStore(catalog.DemoProducts())
	cartStore := cart.NewMemoryStore(catalogStore)
	chatService := chat.NewService(catalogStore, cartStore)

	handler := httpapi.NewHandler(chatService, catalogStore, cartStore)
	server := &http.Server{Addr: ":8080", Handler: handler}

	log.Println("ekt assistant backend listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}