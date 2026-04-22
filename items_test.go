package loyverse_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func TestGetItems_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"items": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	got, err := c.GetItems(context.Background())
	if err != nil {
		t.Fatalf("GetItems() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetItems() returned %d items, want 0", len(got))
	}
}

func TestGetItems_singlePage(t *testing.T) {
	fixture := itemFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"items":  []loyverse.Item{fixture},
			"cursor": "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.GetItems(context.Background())
	if err != nil {
		t.Fatalf("GetItems() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetItems() returned %d items, want 1", len(got))
	}
	if diff := cmp.Diff(fixture.ID, got[0].ID); diff != "" {
		t.Errorf("GetItems()[0].ID mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(fixture.Name, got[0].Name); diff != "" {
		t.Errorf("GetItems()[0].Name mismatch (-want +got):\n%s", diff)
	}
}

func TestGetItems_multiPage(t *testing.T) {
	item1 := itemFixture()
	item2 := loyverse.Item{ID: "item-2", Name: "Second Item"}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			mustWriteJSON(t, w, map[string]any{
				"items":  []loyverse.Item{item1},
				"cursor": "page-2",
			})
		} else {
			mustWriteJSON(t, w, map[string]any{
				"items":  []loyverse.Item{item2},
				"cursor": "",
			})
		}
	})
	c := newTestClient(t, mux)

	got, err := c.GetItems(context.Background())
	if err != nil {
		t.Fatalf("GetItems() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetItems() returned %d items, want 2", len(got))
	}
}

func TestGetItem(t *testing.T) {
	fixture := itemFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, fixture)
	})
	c := newTestClient(t, mux)

	got, err := c.GetItem(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("GetItem() error = %v", err)
	}
	if got.ID != fixture.ID {
		t.Errorf("GetItem().ID = %q, want %q", got.ID, fixture.ID)
	}
	if got.Name != fixture.Name {
		t.Errorf("GetItem().Name = %q, want %q", got.Name, fixture.Name)
	}
}

func TestCreateItem(t *testing.T) {
	fixture := itemFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, fixture)
	})
	c := newTestClient(t, mux)

	req := loyverse.CreateItemRequest{
		Name:       "Test Item",
		TrackStock: true,
		Variants: []loyverse.CreateVariantRequest{
			{PricingType: "FIXED", DefaultPrice: 9.99},
		},
	}
	got, err := c.CreateItem(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}
	if got.ID != fixture.ID {
		t.Errorf("CreateItem().ID = %q, want %q", got.ID, fixture.ID)
	}
}

func TestSetItemCost(t *testing.T) {
	fixture := itemFixture()
	var gotBody map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, fixture)
	})
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("POST /items: decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	const wantCost = 25.0
	if err := c.SetItemCost(context.Background(), "item-1", wantCost); err != nil {
		t.Fatalf("SetItemCost() error = %v", err)
	}

	if gotBody == nil {
		t.Fatal("SetItemCost() did not POST to /items")
	}
	gotCost, ok := gotBody["cost"].(float64)
	if !ok {
		t.Fatalf("POST body[\"cost\"] type = %T, want float64", gotBody["cost"])
	}
	if gotCost != wantCost {
		t.Errorf("SetItemCost() POST body[\"cost\"] = %v, want %v", gotCost, wantCost)
	}
}

func TestResetCategoryPrices(t *testing.T) {
	const catID = "cat-1"
	itemInCat := loyverse.Item{
		ID:         "item-1",
		CategoryID: catID,
		Variants:   []loyverse.Variant{{ID: "var-1", ItemID: "item-1", DefaultPrice: 9.99, PricingType: "FIXED"}},
	}
	itemOutCat := loyverse.Item{
		ID:         "item-2",
		CategoryID: "cat-other",
		Variants:   []loyverse.Variant{{ID: "var-2", ItemID: "item-2"}},
	}

	var mu sync.Mutex
	var postedItemIDs []string
	var postedVariantPrice float64

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"items":  []loyverse.Item{itemInCat, itemOutCat},
			"cursor": "",
		})
	})
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") == itemInCat.ID {
			mustWriteJSON(t, w, itemInCat)
		} else {
			mustWriteJSON(t, w, itemOutCat)
		}
	})
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("POST /items: decode body: %v", err)
		}
		variants := body["variants"].([]any)
		v := variants[0].(map[string]any)
		mu.Lock()
		postedItemIDs = append(postedItemIDs, body["id"].(string))
		postedVariantPrice = v["default_price"].(float64)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	c := newTestClient(t, mux)

	ok, failed, err := c.ResetCategoryPrices(context.Background(), catID)
	if err != nil {
		t.Fatalf("ResetCategoryPrices() error = %v", err)
	}
	if failed != 0 {
		t.Errorf("ResetCategoryPrices() failed = %d, want 0", failed)
	}
	if ok != 1 {
		t.Errorf("ResetCategoryPrices() ok = %d, want 1", ok)
	}
	if len(postedItemIDs) != 1 || postedItemIDs[0] != itemInCat.ID {
		t.Errorf("ResetCategoryPrices() updated items = %v, want only [%s]", postedItemIDs, itemInCat.ID)
	}
	if postedVariantPrice != 0 {
		t.Errorf("ResetCategoryPrices() variant default_price = %v, want 0", postedVariantPrice)
	}
}

func TestResetCategoryPrices_noCategoryItems(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"items": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	ok, failed, err := c.ResetCategoryPrices(context.Background(), "cat-empty")
	if err != nil {
		t.Fatalf("ResetCategoryPrices() error = %v, want nil", err)
	}
	if ok != 0 || failed != 0 {
		t.Errorf("ResetCategoryPrices() = (%d, %d), want (0, 0)", ok, failed)
	}
}

func TestResetCategoryPrices_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		mustWriteJSON(t, w, map[string]any{"code": "ERR", "message": "server error"})
	})
	c := newTestClient(t, mux)

	_, _, err := c.ResetCategoryPrices(context.Background(), "cat-1")
	if err == nil {
		t.Fatal("ResetCategoryPrices() error = nil, want non-nil on API error")
	}
}
