package loyverse_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"testing"

	"github.com/carlospereira5/loyverse"
)

func TestGetInventoryLevels(t *testing.T) {
	level := inventoryLevelFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": []loyverse.InventoryLevel{level},
			"cursor":           "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.GetInventoryLevels(context.Background())
	if err != nil {
		t.Fatalf("GetInventoryLevels() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetInventoryLevels() returned %d levels, want 1", len(got))
	}
	if got[0].VariantID != level.VariantID {
		t.Errorf("GetInventoryLevels()[0].VariantID = %q, want %q", got[0].VariantID, level.VariantID)
	}
	if got[0].InStock != level.InStock {
		t.Errorf("GetInventoryLevels()[0].InStock = %v, want %v", got[0].InStock, level.InStock)
	}
}

func TestGetInventoryLevels_multiPage(t *testing.T) {
	level1 := inventoryLevelFixture()
	level2 := loyverse.InventoryLevel{VariantID: "var-2", StoreID: "store-1", InStock: 5}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			mustWriteJSON(t, w, map[string]any{
				"inventory_levels": []loyverse.InventoryLevel{level1},
				"cursor":           "page-2",
			})
		} else {
			mustWriteJSON(t, w, map[string]any{
				"inventory_levels": []loyverse.InventoryLevel{level2},
				"cursor":           "",
			})
		}
	})
	c := newTestClient(t, mux)

	got, err := c.GetInventoryLevels(context.Background())
	if err != nil {
		t.Fatalf("GetInventoryLevels() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetInventoryLevels() returned %d levels, want 2", len(got))
	}
}

func TestGetItemStock(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, itemFixture())
	})
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": []loyverse.InventoryLevel{inventoryLevelFixture()},
			"cursor":           "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.GetItemStock(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("GetItemStock() error = %v", err)
	}
	if got != 10 {
		t.Errorf("GetItemStock() = %v, want 10", got)
	}
}

func TestGetItemStock_noInventoryRecord(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, itemFixture())
	})
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": []any{},
			"cursor":           "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.GetItemStock(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("GetItemStock() error = %v, want nil", err)
	}
	if got != 0 {
		t.Errorf("GetItemStock() = %v, want 0 when no inventory record", got)
	}
}

func TestSetStock(t *testing.T) {
	var gotBody map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("POST /inventory: decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	if err := c.SetStock(context.Background(), "var-1", "store-1", 42); err != nil {
		t.Fatalf("SetStock() error = %v", err)
	}

	levels, ok := gotBody["inventory_levels"].([]any)
	if !ok || len(levels) != 1 {
		t.Fatalf("POST body inventory_levels = %v, want 1 entry", gotBody["inventory_levels"])
	}
	level := levels[0].(map[string]any)
	if v := level["variant_id"].(string); v != "var-1" {
		t.Errorf("inventory_levels[0].variant_id = %q, want %q", v, "var-1")
	}
	if v := level["stock_after"].(float64); v != 42 {
		t.Errorf("inventory_levels[0].stock_after = %v, want 42", v)
	}
}

func TestAdjustStock(t *testing.T) {
	var gotBody map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, itemFixture())
	})
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": []loyverse.InventoryLevel{inventoryLevelFixture()}, // InStock: 10
			"cursor":           "",
		})
	})
	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("POST /inventory: decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	// Current stock is 10; adding 5 should result in stock_after = 15.
	if err := c.AdjustStock(context.Background(), "item-1", 5); err != nil {
		t.Fatalf("AdjustStock() error = %v", err)
	}

	levels := gotBody["inventory_levels"].([]any)
	level := levels[0].(map[string]any)
	if stockAfter := level["stock_after"].(float64); stockAfter != 15 {
		t.Errorf("AdjustStock(+5) POST stock_after = %v, want 15 (10 + 5)", stockAfter)
	}
}

func TestAdjustStock_subtract(t *testing.T) {
	var gotBody map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, itemFixture())
	})
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": []loyverse.InventoryLevel{inventoryLevelFixture()}, // InStock: 10
			"cursor":           "",
		})
	})
	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	// Current stock is 10; subtracting 3 should result in stock_after = 7.
	if err := c.AdjustStock(context.Background(), "item-1", -3); err != nil {
		t.Fatalf("AdjustStock(-3) error = %v", err)
	}

	levels := gotBody["inventory_levels"].([]any)
	level := levels[0].(map[string]any)
	if stockAfter := level["stock_after"].(float64); stockAfter != 7 {
		t.Errorf("AdjustStock(-3) POST stock_after = %v, want 7 (10 - 3)", stockAfter)
	}
}

func TestUpdateStockBatch(t *testing.T) {
	levels := []loyverse.InventoryLevel{
		{VariantID: "var-1", StoreID: "store-1", InStock: 10},
		{VariantID: "var-2", StoreID: "store-1", InStock: 5},
		{VariantID: "var-3", StoreID: "store-1", InStock: 0},
	}
	updates := map[string]float64{
		"var-1": 20,
		"var-2": 15,
		"var-3": 8,
	}

	var mu sync.Mutex
	var updatedVariants []string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": levels,
			"cursor":           "",
		})
	})
	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("POST /inventory: decode body: %v", err)
		}
		lvls := body["inventory_levels"].([]any)
		for _, l := range lvls {
			entry := l.(map[string]any)
			mu.Lock()
			updatedVariants = append(updatedVariants, entry["variant_id"].(string))
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	ok, failed, err := c.UpdateStockBatch(context.Background(), updates)
	if err != nil {
		t.Fatalf("UpdateStockBatch() error = %v", err)
	}
	if failed != 0 {
		t.Errorf("UpdateStockBatch() failed = %d, want 0", failed)
	}
	if ok != 3 {
		t.Errorf("UpdateStockBatch() ok = %d, want 3", ok)
	}

	sort.Strings(updatedVariants)
	want := []string{"var-1", "var-2", "var-3"}
	for i, v := range want {
		if updatedVariants[i] != v {
			t.Errorf("UpdateStockBatch() updated variants[%d] = %q, want %q", i, updatedVariants[i], v)
		}
	}
}

func TestResetNegativeStock(t *testing.T) {
	levels := []loyverse.InventoryLevel{
		{VariantID: "var-pos", StoreID: "store-1", InStock: 10},  // positive — must NOT be reset
		{VariantID: "var-neg1", StoreID: "store-1", InStock: -5}, // negative — must be reset
		{VariantID: "var-neg2", StoreID: "store-1", InStock: -3}, // negative — must be reset
	}

	var mu sync.Mutex
	var resetVariants []string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"inventory_levels": levels,
			"cursor":           "",
		})
	})
	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("POST /inventory: decode body: %v", err)
		}
		lvls := body["inventory_levels"].([]any)
		for _, l := range lvls {
			entry := l.(map[string]any)
			mu.Lock()
			resetVariants = append(resetVariants, entry["variant_id"].(string))
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	ok, failed, err := c.ResetNegativeStock(context.Background())
	if err != nil {
		t.Fatalf("ResetNegativeStock() error = %v", err)
	}
	if failed != 0 {
		t.Errorf("ResetNegativeStock() failed = %d, want 0", failed)
	}
	if ok != 2 {
		t.Errorf("ResetNegativeStock() ok = %d, want 2 (only negative levels)", ok)
	}

	sort.Strings(resetVariants)
	want := []string{"var-neg1", "var-neg2"}
	for i, v := range want {
		if resetVariants[i] != v {
			t.Errorf("ResetNegativeStock() reset variants[%d] = %q, want %q", i, resetVariants[i], v)
		}
	}

	for _, v := range resetVariants {
		if v == "var-pos" {
			t.Error("ResetNegativeStock() reset a positive stock level, want only negative levels reset")
		}
	}
}

func TestResetCategoryStock(t *testing.T) {
	const catID = "cat-1"
	items := []loyverse.Item{
		{
			ID:         "item-1",
			CategoryID: catID,
			Variants:   []loyverse.Variant{{ID: "var-cat-1", ItemID: "item-1"}, {ID: "var-cat-2", ItemID: "item-1"}},
		},
		{
			ID:         "item-2",
			CategoryID: "cat-other",
			Variants:   []loyverse.Variant{{ID: "var-other", ItemID: "item-2"}},
		},
	}
	levels := []loyverse.InventoryLevel{
		{VariantID: "var-cat-1", StoreID: "store-1", InStock: 15},
		{VariantID: "var-cat-2", StoreID: "store-1", InStock: 8},
		{VariantID: "var-other", StoreID: "store-1", InStock: 5},
	}

	var mu sync.Mutex
	var resetVariants []string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"items": items, "cursor": ""})
	})
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"inventory_levels": levels, "cursor": ""})
	})
	mux.HandleFunc("POST /inventory", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("POST /inventory: decode body: %v", err)
		}
		lvls := body["inventory_levels"].([]any)
		for _, l := range lvls {
			entry := l.(map[string]any)
			mu.Lock()
			resetVariants = append(resetVariants, entry["variant_id"].(string))
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	ok, failed, err := c.ResetCategoryStock(context.Background(), catID)
	if err != nil {
		t.Fatalf("ResetCategoryStock() error = %v", err)
	}
	if failed != 0 {
		t.Errorf("ResetCategoryStock() failed = %d, want 0", failed)
	}
	if ok != 2 {
		t.Errorf("ResetCategoryStock() ok = %d, want 2", ok)
	}

	sort.Strings(resetVariants)
	want := []string{"var-cat-1", "var-cat-2"}
	for i, v := range want {
		if resetVariants[i] != v {
			t.Errorf("ResetCategoryStock() reset variants[%d] = %q, want %q", i, resetVariants[i], v)
		}
	}
	for _, v := range resetVariants {
		if v == "var-other" {
			t.Error("ResetCategoryStock() reset a variant outside the target category")
		}
	}
}

func TestResetCategoryStock_noCategoryItems(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"items": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	ok, failed, err := c.ResetCategoryStock(context.Background(), "cat-empty")
	if err != nil {
		t.Fatalf("ResetCategoryStock() error = %v, want nil", err)
	}
	if ok != 0 || failed != 0 {
		t.Errorf("ResetCategoryStock() = (%d, %d), want (0, 0)", ok, failed)
	}
}

func TestResetCategoryStock_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		mustWriteJSON(t, w, map[string]any{"code": "ERR", "message": "server error"})
	})
	c := newTestClient(t, mux)

	_, _, err := c.ResetCategoryStock(context.Background(), "cat-1")
	if err == nil {
		t.Fatal("ResetCategoryStock() error = nil, want non-nil on API error")
	}
}
