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
		product.Stock = product.Quantity
		product.CertificateURL = ""
		if len(product.CertificateURLs) > 0 {
			product.CertificateURL = product.CertificateURLs[0]
		}
		if product.SKU == "" {
			product.SKU = product.Article
		}
		bySKU[product.SKU] = product
	}
	return &MemoryStore{products: bySKU}
}

func (store *MemoryStore) Search(query string) []Product {
	query = strings.ToLower(strings.TrimSpace(query))
	store.mu.RLock()
	defer store.mu.RUnlock()

	matches := make([]Product, 0)
	exactMatches := make([]Product, 0)
	for _, product := range store.products {
		searchText := strings.ToLower(strings.Join([]string{
			product.SKU,
			product.Article,
			product.SupplierArticle,
			product.Name,
			product.Brand,
			product.Category,
			product.ProductType,
		}, " "))
		if query == "" {
			matches = append(matches, product)
			continue
		}

		candidate := strings.ToLower(product.SKU) == query ||
			strings.ToLower(product.Article) == query ||
			strings.ToLower(product.SupplierArticle) == query
		if candidate {
			exactMatches = append(exactMatches, product)
			continue
		}
		if strings.Contains(searchText, query) {
			matches = append(matches, product)
		}
	}
	if len(exactMatches) > 0 {
		return append(exactMatches, matches...)
	}
	return matches
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
		if product.SKU == source.SKU {
			continue
		}
		if product.Quantity <= 0 && product.Stock <= 0 {
			continue
		}
		if product.Category != source.Category {
			continue
		}

		currentMatch := product.Characteristics.Current != "" && source.Characteristics.Current != "" &&
			strings.EqualFold(strings.TrimSpace(product.Characteristics.Current), strings.TrimSpace(source.Characteristics.Current))
			voltageMatch := product.Characteristics.Voltage != "" && source.Characteristics.Voltage != "" &&
			strings.EqualFold(strings.TrimSpace(product.Characteristics.Voltage), strings.TrimSpace(source.Characteristics.Voltage))
			polesMatch := product.Characteristics.Poles != "" && source.Characteristics.Poles != "" &&
			strings.EqualFold(strings.TrimSpace(product.Characteristics.Poles), strings.TrimSpace(source.Characteristics.Poles))

		if currentMatch || (voltageMatch && polesMatch) {
			result = append(result, product)
		}
	}
	return result
}

func DemoProducts() []Product {
	first := Product{
		ID:              515291,
		Article:         "200300285_",
		SupplierArticle: "027228",
		SKU:             "ABB-S201-C16",
		Name:            "Автоматический выключатель ABB S201 C16",
		Description:     "Автоматический выключатель DRX250 MT 3P 160А 18kA (арт. 027228) Legrand – мощное защитное устройство для предотвращения перегрузок и коротких замыканий в электрических сетях.",
		Price:           4200,
		Quantity:        12,
		Brand:           "ABB",
		ProductType:     "Автоматический выключатель",
		Category:        "Автоматы",
		Characteristics: ProductCharacteristics{
			Poles:            "1",
			Current:          "16 A",
			Voltage:          "230В",
			BreakingCapacity: "6кА",
			InstallationType: "Винтовое",
		},
		RelatedProductIDs: []int{48783, 23466, 28727},
		MinimumMultiple:   1,
		Image:             "https://ekt.kz/example/abb-s201-c16.jpg",
		URL:               "https://ekt.kz/catalog/abb-s201-c16",
		DataIssues:        []string{"Номинальный ток из карточки требует проверки; значения потенциально расходятся с описанием."},
		Stock:            12,
		AvailableStores: []StoreInfo{{ID: 1, Name: "Алматы", Quantity: 12}},
	}

	second := Product{
		ID:              1001,
		Article:         "EKF-BA-16",
		SupplierArticle: "EKF-BA-16",
		SKU:             "EKF-BA-16",
		Name:            "Автоматический выключатель EKF BA 47-29 C16",
		Description:     "Автоматический выключатель EKF BA 47-29 C16 для стандартных электросетей.",
		Category:        "Автоматы",
		Price:           1850,
		Quantity:        0,
		Brand:           "EKF",
		ProductType:     "Автоматический выключатель",
		Characteristics: ProductCharacteristics{
			Poles:   "1",
			Current: "16 A",
			Voltage: "220В",
		},
		Stock: 0,
	}

	third := Product{
		ID:              1002,
		Article:         "IEK-C25",
		SupplierArticle: "IEK-C25",
		SKU:             "IEK-C25",
		Name:            "Автоматический выключатель IEK C25",
		Description:     "Автоматический выключатель IEK C25 для силовых нагрузок.",
		Category:        "Автоматы",
		Price:           2100,
		Quantity:        7,
		Brand:           "IEK",
		ProductType:     "Автоматический выключатель",
		Characteristics: ProductCharacteristics{
			Poles:   "1",
			Current: "25 A",
			Voltage: "220В",
		},
		Stock: 7,
	}

	return []Product{first, second, third}
}