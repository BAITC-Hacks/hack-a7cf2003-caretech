package main

import (
	"log"
	"net/http"

	"hackalem-ekt-backend/internal/cart"
	"hackalem-ekt-backend/internal/catalog"
	"hackalem-ekt-backend/internal/chat"
	db "hackalem-ekt-backend/internal/db"
	httpapi "hackalem-ekt-backend/internal/httpapi"
	"hackalem-ekt-backend/internal/llm"
)

func main() {
	catalogStore := catalog.NewMemoryStore(catalog.DemoProducts())
	cartStore := cart.NewMemoryStore(catalogStore)

	if conn, err := db.NewFromEnv(); err != nil {
		log.Printf("PostgreSQL unavailable, fallback to in-memory mode: %v", err)
	} else if conn != nil {
		if err := db.EnsureSchema(conn); err != nil {
			log.Printf("PostgreSQL schema init failed: %v", err)
		} else {
			log.Println("PostgreSQL connected and ready")
		}
		_ = conn.Close()
	}

	var opts []chat.Option
	if client := llm.NewClientFromEnv(); client != nil {
		opts = append(opts, chat.WithLLMClient(client))
		log.Println("NVIDIA LLM client initialized")
	} else {
		log.Println("NVIDIA LLM client not configured; running catalog-first fallback mode")
	}

	chatService := chat.NewService(catalogStore, cartStore, opts...)

	handler := httpapi.NewHandler(chatService, catalogStore, cartStore)
	server := &http.Server{Addr: ":8080", Handler: handler}

	log.Println("ekt assistant backend listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}