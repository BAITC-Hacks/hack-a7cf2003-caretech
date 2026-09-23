package chat

import (
	"strings"
	"testing"

	"hackalem-ekt-backend/internal/cart"
	"hackalem-ekt-backend/internal/catalog"
)

func newTestService() (*Service, *cart.MemoryStore) {
	products := catalog.NewMemoryStore(catalog.DemoProducts())
	carts := cart.NewMemoryStore(products)
	return NewService(products, carts), carts
}

func TestHandleDoesNotChangeCartBeforeConfirmation(t *testing.T) {
	service, carts := newTestService()
	response := service.Handle(Request{SessionID: "test", Message: "Есть ABB-S201-C16?"})

	if response.PendingAdd == nil {
		t.Fatal("expected pending add")
	}
	if items := carts.Get("test").Items; len(items) != 0 {
		t.Fatalf("cart changed before confirmation: %+v", items)
	}
}

func TestHandleSuggestsAlternativeForOutOfStockProduct(t *testing.T) {
	service, _ := newTestService()
	response := service.Handle(Request{SessionID: "test", Message: "Есть EKF-BA-16?"})

	if len(response.Products) == 0 {
		t.Fatal("expected at least one alternative")
	}
	if !strings.Contains(response.Reply, "аналог") {
		t.Fatalf("unexpected reply: %s", response.Reply)
	}
}

func TestHandleConfirmAddsPendingProductToCart(t *testing.T) {
	service, carts := newTestService()
	service.Handle(Request{SessionID: "test", Message: "Есть ABB-S201-C16?"})
	response := service.Handle(Request{SessionID: "test", Message: "Да, добавь"})

	if !strings.Contains(response.Reply, "Готово") {
		t.Fatalf("expected success confirmation, got %q", response.Reply)
	}
	if items := carts.Get("test").Items; len(items) != 1 {
		t.Fatalf("expected one item in cart after confirmation, got %+v", items)
	}
}

func TestHandleRejectsConfirmationWithoutPendingProduct(t *testing.T) {
	service, _ := newTestService()
	response := service.Handle(Request{SessionID: "test", Message: "Да, добавь"})

	if !strings.Contains(response.Reply, "нет ожидающего") && !strings.Contains(response.Reply, "ожидающего") {
		t.Fatalf("expected no-pending confirmation reply, got %q", response.Reply)
	}
}

func TestNormalizeQueryPreservesTrailingUnderscoreInArticle(t *testing.T) {
	query := normalizeQuery("Есть 200300285_?")
	if query != "200300285_" {
		t.Fatalf("expected article identifier to keep trailing underscore, got %q", query)
	}
}

func TestParseStructuredReply(t *testing.T) {
	reply := parseStructuredReply(`{"intent":"other","reply":"Уточните артикул.","product_ids":[],"alternative_ids":[],"requested_action":null,"needs_clarification":true,"needs_human":false}`)
	if reply != "Уточните артикул." {
		t.Fatalf("unexpected parsed reply: %q", reply)
	}
}
