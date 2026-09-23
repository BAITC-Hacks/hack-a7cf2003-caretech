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