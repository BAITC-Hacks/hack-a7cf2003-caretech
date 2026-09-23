package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ektProductsPage struct {
	Items []ektProduct `json:"items"`
}

type ektProduct struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Article     string            `json:"article"`
	Price       int               `json:"price"`
	Quantity    *int              `json:"quantity"`
	Image       *string           `json:"image"`
	URL         string            `json:"url"`
	DetailURL   string            `json:"url_api_detail"`
	Properties  map[string]string `json:"properties"`
	DataIssues  []string          `json:"data_issues"`
}

// LoadEKTJSON reads one or more page exports. Later files replace duplicate IDs.
func LoadEKTJSON(paths ...string) ([]Product, error) {
	byID := make(map[int]Product)
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, fmt.Errorf("read EKT catalog %q: %w", path, err)
		}

		items, err := parseEKTProducts(data)
		if err != nil {
			return nil, fmt.Errorf("parse EKT catalog %q: %w", path, err)
		}
		for _, item := range items {
			quantity := 0
			if item.Quantity != nil {
				quantity = *item.Quantity
			}
			productType := item.Properties["OBYEM"]
			product := Product{
				ID:           item.ID,
				Article:      item.Article,
				SKU:          item.Article,
				Name:         item.Name,
				Description:  item.Description,
				Price:        item.Price,
				Quantity:     quantity,
				Brand:        item.Properties["TORGOVAYA_MARKA"],
				ProductType:  productType,
				Properties:   item.Properties,
				URL:          item.URL,
				DetailAPIURL: item.DetailURL,
				DataIssues:   item.DataIssues,
				Characteristics: ProductCharacteristics{
					Poles:            item.Properties["KOLICHESTVO_POLYUSOV"],
					Current:          item.Properties["NOMINALNYY_TOK"],
					Voltage:          item.Properties["NOMINALNOE_NAPRYAZHENIE"],
					BreakingCapacity: item.Properties["NOMINALNAYA_OTKLYUCHAYUSHCHAYA_SPOSOBNOST"],
				},
			}
			product.SupplierArticle = item.Properties["ARTIKULPOSTAVSHCHIKA"]
			if product.Description == "" {
				product.Description = product.Name
			}
			if len(product.DataIssues) == 0 && item.Quantity == nil {
				product.DataIssues = []string{"Количество и склады неизвестны: исходный список их не содержит."}
			}
			if item.Image != nil {
				product.Image = *item.Image
			}
			if product.URL == "" {
				product.URL = item.DetailURL
			}
			byID[item.ID] = NormalizeProduct(product)
		}
	}

	products := make([]Product, 0, len(byID))
	for _, product := range byID {
		products = append(products, product)
	}
	return products, nil
}

func parseEKTProducts(data []byte) ([]ektProduct, error) {
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "[") {
		var items []ektProduct
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err
		}
		return items, nil
	}
	var page ektProductsPage
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}
