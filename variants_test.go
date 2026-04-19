package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func variantFixture() loyverse.Variant {
	ts := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
	price := 9.99
	return loyverse.Variant{
		ID:           "var-1",
		ItemID:       "item-1",
		SKU:          "SKU-001",
		Barcode:      "1234567890",
		Cost:         5.00,
		PurchaseCost: 4.50,
		DefaultPrice: price,
		PricingType:  "FIXED",
		Stores: []loyverse.VariantStore{
			{
				StoreID:          "store-1",
				PricingType:      "FIXED",
				Price:            &price,
				AvailableForSale: true,
			},
		},
		CreatedAt: ts,
		UpdatedAt: ts,
	}
}

func TestListVariants(t *testing.T) {
	v := variantFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"variants": []loyverse.Variant{v},
			"cursor":   "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.ListVariants(context.Background(), loyverse.VariantsFilter{})
	if err != nil {
		t.Fatalf("ListVariants() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListVariants() returned %d variants, want 1", len(got))
	}
	if diff := cmp.Diff(v, got[0]); diff != "" {
		t.Errorf("ListVariants()[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestListVariants_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"variants": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	got, err := c.ListVariants(context.Background(), loyverse.VariantsFilter{})
	if err != nil {
		t.Fatalf("ListVariants() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListVariants() returned %d variants, want 0", len(got))
	}
}

func TestListVariants_multiPage(t *testing.T) {
	v1 := variantFixture()
	v2 := loyverse.Variant{ID: "var-2", ItemID: "item-1", SKU: "SKU-002"}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			mustWriteJSON(t, w, map[string]any{
				"variants": []loyverse.Variant{v1},
				"cursor":   "page-2",
			})
		} else {
			mustWriteJSON(t, w, map[string]any{
				"variants": []loyverse.Variant{v2},
				"cursor":   "",
			})
		}
	})
	c := newTestClient(t, mux)

	got, err := c.ListVariants(context.Background(), loyverse.VariantsFilter{})
	if err != nil {
		t.Fatalf("ListVariants() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListVariants() returned %d variants, want 2", len(got))
	}
}

func TestListVariants_filterBySKU(t *testing.T) {
	var gotSKU string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		gotSKU = r.URL.Query().Get("sku")
		mustWriteJSON(t, w, map[string]any{"variants": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	_, _ = c.ListVariants(context.Background(), loyverse.VariantsFilter{SKU: "SKU-001"})

	if gotSKU != "SKU-001" {
		t.Errorf("ListVariants() sent sku = %q, want %q", gotSKU, "SKU-001")
	}
}

func TestListVariants_filterByItemIDs(t *testing.T) {
	var gotItemIDs string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		gotItemIDs = r.URL.Query().Get("items_ids")
		mustWriteJSON(t, w, map[string]any{"variants": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	_, _ = c.ListVariants(context.Background(), loyverse.VariantsFilter{ItemIDs: "item-1,item-2"})

	if gotItemIDs != "item-1,item-2" {
		t.Errorf("ListVariants() sent items_ids = %q, want %q", gotItemIDs, "item-1,item-2")
	}
}

func TestListVariants_filterByVariantIDs(t *testing.T) {
	var gotVariantIDs string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		gotVariantIDs = r.URL.Query().Get("variants_ids")
		mustWriteJSON(t, w, map[string]any{"variants": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	_, _ = c.ListVariants(context.Background(), loyverse.VariantsFilter{VariantIDs: "var-1,var-2"})

	if gotVariantIDs != "var-1,var-2" {
		t.Errorf("ListVariants() sent variants_ids = %q, want %q", gotVariantIDs, "var-1,var-2")
	}
}

func TestListVariants_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	c := newTestClient(t, mux)

	_, err := c.ListVariants(context.Background(), loyverse.VariantsFilter{})
	if err == nil {
		t.Fatal("ListVariants() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetVariant(t *testing.T) {
	v := variantFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, v)
	})
	c := newTestClient(t, mux)

	got, err := c.GetVariant(context.Background(), "var-1")
	if err != nil {
		t.Fatalf("GetVariant() error = %v", err)
	}
	if diff := cmp.Diff(v, *got); diff != "" {
		t.Errorf("GetVariant() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetVariant_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /variants/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetVariant(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetVariant() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
