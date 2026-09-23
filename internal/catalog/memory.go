package catalog

import (
	"sort"
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
	query = strings.TrimSpace(query)
	if query == "" {
		store.mu.RLock()
		defer store.mu.RUnlock()
		products := make([]Product, 0, len(store.products))
		for _, product := range store.products {
			products = append(products, product)
		}
		sort.Slice(products, func(i, j int) bool {
			if products[i].Stock != products[j].Stock {
				return products[i].Stock > products[j].Stock
			}
			return products[i].Name < products[j].Name
		})
		return products
	}

	queryLower := strings.ToLower(query)
	store.mu.RLock()
	defer store.mu.RUnlock()

	results := make([]Product, 0, len(store.products))
	for _, product := range store.products {
		score := productMatchScore(product, queryLower)
		if score > 0 {
			results = append(results, product)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		leftScore := productMatchScore(results[i], queryLower)
		rightScore := productMatchScore(results[j], queryLower)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		if results[i].Stock != results[j].Stock {
			return results[i].Stock > results[j].Stock
		}
		return results[i].Name < results[j].Name
	})
	return results
}

func productMatchScore(product Product, query string) int {
	if query == "" {
		return 0
	}

	if strings.EqualFold(product.SKU, query) ||
		strings.EqualFold(product.Article, query) ||
		strings.EqualFold(product.SupplierArticle, query) {
		return 100
	}

	searchText := strings.ToLower(strings.Join([]string{
		product.SKU,
		product.Article,
		product.SupplierArticle,
		product.Name,
		product.Brand,
		product.Category,
		product.ProductType,
	}, " "))
	if strings.Contains(searchText, query) {
		if strings.Contains(strings.ToLower(product.SKU), query) ||
			strings.Contains(strings.ToLower(product.Article), query) ||
			strings.Contains(strings.ToLower(product.SupplierArticle), query) {
			return 90
		}
		if strings.Contains(strings.ToLower(product.Name), query) {
			return 70
		}
		return 50
	}
	return 0
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

	candidates := make([]struct {
		product Product
		score   int
	}, 0)

	for _, product := range store.products {
		if product.SKU == source.SKU {
			continue
		}
		if product.Quantity <= 0 && product.Stock <= 0 {
			continue
		}
		if source.Category != "" && product.Category != "" && source.Category != product.Category {
			continue
		}

		score := 0
		if strings.EqualFold(strings.TrimSpace(product.Category), strings.TrimSpace(source.Category)) {
			score += 1
		}
		if product.ProductType != "" && source.ProductType != "" &&
			strings.EqualFold(strings.TrimSpace(product.ProductType), strings.TrimSpace(source.ProductType)) {
			score += 1
		}
		if product.Characteristics.Current != "" && source.Characteristics.Current != "" &&
			strings.EqualFold(strings.TrimSpace(product.Characteristics.Current), strings.TrimSpace(source.Characteristics.Current)) {
			score += 5
		}
		if product.Characteristics.Voltage != "" && source.Characteristics.Voltage != "" &&
			strings.EqualFold(strings.TrimSpace(product.Characteristics.Voltage), strings.TrimSpace(source.Characteristics.Voltage)) {
			score += 3
		}
		if product.Characteristics.Poles != "" && source.Characteristics.Poles != "" &&
			strings.EqualFold(strings.TrimSpace(product.Characteristics.Poles), strings.TrimSpace(source.Characteristics.Poles)) {
			score += 2
		}
		if score > 0 {
			candidates = append(candidates, struct {
				product Product
				score   int
			}{product: product, score: score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		if candidates[i].product.Stock != candidates[j].product.Stock {
			return candidates[i].product.Stock > candidates[j].product.Stock
		}
		return candidates[i].product.Name < candidates[j].product.Name
	})

	result := make([]Product, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, candidate.product)
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