package catalog

import "strings"

type ProductCharacteristics struct {
	Poles            string `json:"poles,omitempty"`
	Current          string `json:"current,omitempty"`
	Voltage          string `json:"voltage,omitempty"`
	BreakingCapacity string `json:"breaking_capacity,omitempty"`
	InstallationType string `json:"installation_type,omitempty"`
}

type StoreInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type Product struct {
	ID                int                     `json:"id"`
	Article           string                  `json:"article,omitempty"`
	SupplierArticle   string                  `json:"supplier_article,omitempty"`
	Name              string                  `json:"name"`
	Description       string                  `json:"description,omitempty"`
	Price             int                     `json:"price"`
	Quantity          int                     `json:"quantity"`
	AvailableStores   []StoreInfo             `json:"available_stores,omitempty"`
	Brand             string                  `json:"brand,omitempty"`
	ProductType       string                  `json:"product_type,omitempty"`
	Category          string                  `json:"category,omitempty"`
	Characteristics   ProductCharacteristics  `json:"characteristics"`
	RelatedProductIDs []int                   `json:"related_product_ids,omitempty"`
	MinimumMultiple   int                     `json:"minimum_multiple,omitempty"`
	CertificateURLs   []string                `json:"certificate_urls,omitempty"`
	Image             string                  `json:"image,omitempty"`
	URL               string                  `json:"url,omitempty"`
	DataIssues        []string                `json:"data_issues,omitempty"`

	SKU            string `json:"sku,omitempty"`
	Stock          int    `json:"stock,omitempty"`
	CertificateURL string `json:"certificate_url,omitempty"`
}

type Repository interface {
	Search(query string) []Product
	FindBySKU(sku string) (Product, bool)
	Alternatives(product Product) []Product
}

func NormalizeProduct(product Product) Product {
	if product.SKU == "" && product.Article != "" {
		product.SKU = product.Article
	}
	if product.Article == "" && product.SKU != "" {
		product.Article = product.SKU
	}

	if product.Characteristics.Current != "" && product.Name != "" {
		if strings.Contains(strings.ToLower(product.Name), "160") && strings.Contains(strings.ToLower(product.Characteristics.Current), "250") {
			product.Characteristics.Current = ""
			product.DataIssues = append(product.DataIssues, "Номинальный ток: name и description = 160 А; properties.NOMINALNYY_TOK = 250 А. Значение требует проверки.")
		}
	}

	if product.Quantity < 0 {
		product.Quantity = 0
	}
	if product.Stock < 0 {
		product.Stock = product.Quantity
	}
	if product.Stock == 0 && product.Quantity > 0 {
		product.Stock = product.Quantity
	}
	return product
}