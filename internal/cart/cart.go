package cart

import (
	"fmt"
	"sync"

	"hackalem-ekt-backend/internal/catalog"
)

type Item struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}

type Cart struct {
	Items []Item `json:"items"`
}

type Store interface {
	Get(sessionID string) Cart
	Add(sessionID, sku string, quantity int) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	catalog catalog.Repository
	carts   map[string]Cart
}

func NewMemoryStore(products catalog.Repository) *MemoryStore {
	return &MemoryStore{catalog: products, carts: make(map[string]Cart)}
}

func (store *MemoryStore) Get(sessionID string) Cart {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.carts[sessionID]
}

func (store *MemoryStore) Add(sessionID, sku string, quantity int) error {
	product, ok := store.catalog.FindBySKU(sku)
	if !ok {
		return fmt.Errorf("товар не найден")
	}
	if quantity < 1 || quantity > product.Stock {
		return fmt.Errorf("количество превышает остаток: доступно %d", product.Stock)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	cart := store.carts[sessionID]
	cart.Items = append(cart.Items, Item{SKU: sku, Name: product.Name, Quantity: quantity, Price: product.Price})
	store.carts[sessionID] = cart
	return nil
}