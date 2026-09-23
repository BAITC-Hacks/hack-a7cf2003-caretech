package catalog

type Product struct {
	SKU             string            `json:"sku"`
	Name            string            `json:"name"`
	Category        string            `json:"category"`
	Price           int               `json:"price"`
	Stock           int               `json:"stock"`
	Characteristics map[string]string `json:"characteristics"`
	CertificateURL  string            `json:"certificate_url,omitempty"`
}

type Repository interface {
	Search(query string) []Product
	FindBySKU(sku string) (Product, bool)
	Alternatives(product Product) []Product
}