package catalog

import "testing"

func TestDemoProductMatchesEKTContract(t *testing.T) {
	products := DemoProducts()
	if len(products) == 0 {
		t.Fatal("expected demo products")
	}

	product := products[0]
	if product.ID == 0 {
		t.Fatal("expected EKT product id")
	}
	if product.Article == "" {
		t.Fatal("expected article to be preserved")
	}
	if product.Quantity != product.Stock {
		t.Fatalf("quantity and stock must stay aligned: quantity=%d stock=%d", product.Quantity, product.Stock)
	}
	if product.Name == "" {
		t.Fatal("expected product name")
	}
	if product.Brand == "" {
		t.Fatal("expected product brand")
	}
	if len(product.DataIssues) == 0 {
		t.Fatal("expected data issue note for data-quality checks")
	}
}

func TestNormalizeProductPreservesArticleAndRecordsConflicts(t *testing.T) {
	product := Product{
		SKU:     "200300285_",
		Article: "200300285_",
		Name:    "027228 АВ DRX250 MT 3ф 160А",
		Characteristics: ProductCharacteristics{
			Current: "250 A",
		},
	}

	normalized := NormalizeProduct(product)
	if normalized.SKU != "200300285_" {
		t.Fatalf("expected article to keep trailing underscore, got %q", normalized.SKU)
	}
	if normalized.Characteristics.Current != "" && normalized.Characteristics.Current != "250 A" {
		t.Fatalf("unexpected current value after normalization: %q", normalized.Characteristics.Current)
	}
	if len(normalized.DataIssues) == 0 {
		t.Fatal("expected data issue for conflicting current values")
	}
}

func TestSearchPrioritizesExactArticleAndSupplierArticleMatches(t *testing.T) {
	store := NewMemoryStore(DemoProducts())
	results := store.Search("027228")
	if len(results) == 0 {
		t.Fatal("expected search result")
	}
	if results[0].SupplierArticle != "027228" {
		t.Fatalf("expected supplier article match ranked first, got %q", results[0].SupplierArticle)
	}

	results = store.Search("200300285_")
	if len(results) == 0 {
		t.Fatal("expected search result for article")
	}
	if results[0].Article != "200300285_" {
		t.Fatalf("expected article match ranked first, got %q", results[0].Article)
	}
}
