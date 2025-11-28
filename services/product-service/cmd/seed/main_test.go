package main

import (
	"strings"
	"testing"
)

func TestRandomStringUsesExpectedCharset(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	got := randomString(12)

	if len(got) != 12 {
		t.Fatalf("expected length 12, got %d", len(got))
	}

	for _, ch := range got {
		if !strings.ContainsRune(charset, ch) {
			t.Fatalf("unexpected character %q in random string", ch)
		}
	}
}

func TestRandomRatingRange(t *testing.T) {
	for i := 0; i < 50; i++ {
		r := randomRating()
		if r < 3.5 || r > 5.0 {
			t.Fatalf("rating out of range: %f", r)
		}
	}
}

func TestRandomQuantityRange(t *testing.T) {
	for i := 0; i < 50; i++ {
		q := randomQuantity()
		if q < 10 || q > 109 {
			t.Fatalf("quantity out of range: %d", q)
		}
	}
}

func TestGenerateDimensionsFallback(t *testing.T) {
	dims := generateDimensions("unknown-category")
	if dims.Height != 10 || dims.Width != 10 || dims.Length != 10 || dims.Weight != 1.0 {
		t.Fatalf("expected fallback dimensions, got %+v", dims)
	}
}

func TestGenerateColors(t *testing.T) {
	known := generateColors("Brinquedos")
	if len(known) == 0 || len(known) > 5 {
		t.Fatalf("unexpected number of colors for known category: %d", len(known))
	}

	unknown := generateColors("Unknown")
	if len(unknown) != 1 || unknown[0] != "Variadas" {
		t.Fatalf("expected fallback colors, got %v", unknown)
	}
}

func TestGenerateImagesReturnsThreeEntries(t *testing.T) {
	images := generateImages("Alimentação", "Ração Premium")
	if len(images) != 3 {
		t.Fatalf("expected 3 images, got %d", len(images))
	}
	for _, img := range images {
		if !strings.Contains(img, "images.unsplash.com") {
			t.Fatalf("unexpected image url: %s", img)
		}
	}
}

func TestGeneratePetProducts(t *testing.T) {
	products := generatePetProducts()
	if len(products) != 16 {
		t.Fatalf("expected 16 products, got %d", len(products))
	}

	for _, p := range products {
		if p.Name == "" || p.SKU == "" {
			t.Fatalf("product missing required fields: %+v", p)
		}
		if p.Price <= 0 {
			t.Fatalf("product price should be positive, got %f", p.Price)
		}
		if p.Quantity <= 0 {
			t.Fatalf("product quantity should be positive, got %d", p.Quantity)
		}
		if len(p.Images) != 3 {
			t.Fatalf("expected 3 images for product %s, got %d", p.Name, len(p.Images))
		}
	}
}
