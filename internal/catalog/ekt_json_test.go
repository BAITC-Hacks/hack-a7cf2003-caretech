package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEKTJSONDeduplicatesByID(t *testing.T) {
	directory := t.TempDir()
	first := filepath.Join(directory, "page-1.json")
	second := filepath.Join(directory, "page-2.json")
	if err := os.WriteFile(first, []byte(`{"items":[{"id":7,"name":"Old","article":"A-7","price":10,"url":"old"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte(`{"items":[{"id":7,"name":"New","article":"A-7","price":20,"url":"new"},{"id":8,"name":"Other","article":"A-8","price":30}]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	products, err := LoadEKTJSON(first, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Fatalf("got %d products, want 2", len(products))
	}
	store := NewMemoryStore(products)
	results := store.Search("A-7")
	if len(results) != 1 || results[0].Name != "New" || results[0].Price != 20 {
		t.Fatalf("unexpected exact match: %#v", results)
	}
}

func TestLoadEKTJSONDetailedArray(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "detailed.json")
	content := `[{"id":9,"name":"Автомат 3P 40A","description":"Подробное описание","article":"A-9","price":100,"quantity":4,"url_api_detail":"https://ekt.kz/api/products/detail?id=9","properties":{"OBYEM":"Автоматический выключатель","TORGOVAYA_MARKA":"Legrand","KOLICHESTVO_POLYUSOV":"3","NOMINALNYY_TOK":"40 А","NOMINALNAYA_OTKLYUCHAYUSHCHAYA_SPOSOBNOST":"10 кА","ARTIKULPOSTAVSHCHIKA":"SUP-9"},"data_issues":["Проверить данные"]}]`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	products, err := LoadEKTJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}
	product := products[0]
	if product.Brand != "Legrand" || product.ProductType != "Автоматический выключатель" || product.Quantity != 4 {
		t.Fatalf("unexpected product metadata: %#v", product)
	}
	if product.Characteristics.Poles != "3" || product.Characteristics.Current != "40 А" || product.SupplierArticle != "SUP-9" {
		t.Fatalf("unexpected product characteristics: %#v", product)
	}
}
