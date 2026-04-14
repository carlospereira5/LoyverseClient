package loyverse_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carlospereira5/loyverse"
)

// newTestClient starts an httptest.Server backed by mux and returns a Client pointed at it.
// The server is shut down automatically when the test ends via t.Cleanup.
func newTestClient(t *testing.T, mux *http.ServeMux) *loyverse.Client {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c, err := loyverse.New("test-token", loyverse.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("loyverse.New() error = %v", err)
	}
	return c
}

// mustWriteJSON encodes v as JSON and writes it to w with the correct Content-Type header.
func mustWriteJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("mustWriteJSON: %v", err)
	}
}

// --- fixtures ---

func itemFixture() loyverse.Item {
	return loyverse.Item{
		ID:   "item-1",
		Name: "Test Item",
		Variants: []loyverse.Variant{
			{ID: "var-1", ItemID: "item-1", Barcode: "123456"},
		},
		Stores: []loyverse.ItemStore{
			{StoreID: "store-1"},
		},
	}
}

func inventoryLevelFixture() loyverse.InventoryLevel {
	return loyverse.InventoryLevel{
		VariantID: "var-1",
		StoreID:   "store-1",
		InStock:   10,
	}
}

func receiptFixture() loyverse.Receipt {
	ts := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	return loyverse.Receipt{
		ReceiptNumber: "R-001",
		ReceiptType:   "SALE",
		Status:        "DONE",
		TotalMoney:    100.0,
		ReceiptDate:   ts,
		CreatedAt:     ts,
		LineItems: []loyverse.LineItem{
			{ItemID: "item-1", ItemName: "Test Item", Quantity: 2, Price: 50.0},
		},
	}
}

func categoryFixture() loyverse.Category {
	return loyverse.Category{
		ID:   "cat-1",
		Name: "Beverages",
	}
}

func shiftFixture() loyverse.Shift {
	opened := time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC)
	closed := time.Date(2025, 1, 15, 20, 0, 0, 0, time.UTC)
	return loyverse.Shift{
		ID:       "shift-1",
		OpenedAt: opened,
		ClosedAt: &closed,
		PaidOut:  500.0,
	}
}
