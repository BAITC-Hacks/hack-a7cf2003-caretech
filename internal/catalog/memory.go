package catalog

import (
	"strings"
	"sync"
)

type MemoryStore struct {
	mu       sync.RWMutex
	products map[string]Product
}

func NewMemoryStore(products []Product) *MemoryStore {
	bySKU := make(map[string]Product, len(products))
	for _, product := range products {
		bySKU[product.SKU] = product
	}
	return &MemoryStore{products: bySKU}
}

func (store *MemoryStore) Search(query string) []Product {
	query = strings.ToLower(strings.TrimSpace(query))
	store.mu.RLock()
	defer store.mu.RUnlock()

	result := make([]Product, 0)
	for _, product := range store.products {
		searchText := strings.ToLower(product.SKU + " " + product.Name + " " + product.Category)
		if query == "" || strings.Contains(searchText, query) {
			result = append(result, product)
		}
	}
	return result
}

func (store *MemoryStore) FindBySKU(sku string) (Product, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	product, ok := store.products[sku]
	return product, ok
}

func (store *MemoryStore) Alternatives(source Product) []Product {
	store.mu.RLock()
	defer store.mu.RUnlock()

	result := make([]Product, 0)
	for _, product := range store.products {
		if product.SKU != source.SKU && product.Stock > 0 && product.Category == source.Category && product.Characteristics["ток"] == source.Characteristics["ток"] {
			result = append(result, product)
		}
	}
	return result
}

func DemoProducts() []Product {
	return []Product{
		{SKU: "ABB-S201-C16", Name: "Автоматический выключатель ABB S201 C16", Category: "Автоматы", Price: 4200, Stock: 12, Characteristics: map[string]string{"полюса": "1P", "ток": "16 A", "характеристика": "C"}, CertificateURL: "https://ekt.kz/certificates/abb-s201-c16.pdf"},
		{SKU: "EKF-BA-16", Name: "Автоматический выключатель EKF BA 47-29 C16", Category: "Автоматы", Price: 1850, Stock: 0, Characteristics: map[string]string{"полюса": "1P", "ток": "16 A", "характеристика": "C"}},
		{SKU: "IEK-C25", Name: "Автоматический выключатель IEK C25", Category: "Автоматы", Price: 2100, Stock: 7, Characteristics: map[string]string{"полюса": "1P", "ток": "25 A", "характеристика": "C"}},
	}
}