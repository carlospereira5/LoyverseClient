package loyverse_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/carlospereira5/loyverse"
)

func TestGetReceipt(t *testing.T) {
	receipt := receiptFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /receipts/{number}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, receipt)
	})
	c := newTestClient(t, mux)

	got, err := c.GetReceipt(context.Background(), receipt.ReceiptNumber)
	if err != nil {
		t.Fatalf("GetReceipt() error = %v", err)
	}
	if got.ReceiptNumber != receipt.ReceiptNumber {
		t.Errorf("GetReceipt().ReceiptNumber = %q, want %q", got.ReceiptNumber, receipt.ReceiptNumber)
	}
}

func TestGetReceipt_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /receipts/{number}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		mustWriteJSON(t, w, map[string]any{"errors": []any{}})
	})
	c := newTestClient(t, mux)

	_, err := c.GetReceipt(context.Background(), "R-NOTFOUND")
	if err == nil {
		t.Fatal("GetReceipt() expected error, got nil")
	}
}

func TestCreateReceipt(t *testing.T) {
	receipt := receiptFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /receipts", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, receipt)
	})
	c := newTestClient(t, mux)

	req := loyverse.CreateReceiptRequest{
		StoreID: "store-1",
		LineItems: []loyverse.CreateReceiptLineItem{
			{VariantID: "var-1", Quantity: 2, Price: 50.0},
		},
	}
	got, err := c.CreateReceipt(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateReceipt() error = %v", err)
	}
	if got.ReceiptNumber != receipt.ReceiptNumber {
		t.Errorf("CreateReceipt().ReceiptNumber = %q, want %q", got.ReceiptNumber, receipt.ReceiptNumber)
	}
}

func TestCreateReceipt_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /receipts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		mustWriteJSON(t, w, map[string]any{"errors": []any{}})
	})
	c := newTestClient(t, mux)

	_, err := c.CreateReceipt(context.Background(), loyverse.CreateReceiptRequest{StoreID: "store-1"})
	if err == nil {
		t.Fatal("CreateReceipt() expected error, got nil")
	}
}

func TestListReceipts(t *testing.T) {
	receipt := receiptFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /receipts", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"receipts": []loyverse.Receipt{receipt},
			"cursor":   "",
		})
	})
	c := newTestClient(t, mux)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	got, err := c.ListReceipts(context.Background(), since, until)
	if err != nil {
		t.Fatalf("ListReceipts() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListReceipts() returned %d receipts, want 1", len(got))
	}
	if got[0].ReceiptNumber != receipt.ReceiptNumber {
		t.Errorf("ListReceipts()[0].ReceiptNumber = %q, want %q", got[0].ReceiptNumber, receipt.ReceiptNumber)
	}
}

func TestListReceipts_passesDateRangeParams(t *testing.T) {
	var gotMin, gotMax string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /receipts", func(w http.ResponseWriter, r *http.Request) {
		gotMin = r.URL.Query().Get("created_at_min")
		gotMax = r.URL.Query().Get("created_at_max")
		mustWriteJSON(t, w, map[string]any{"receipts": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	_, _ = c.ListReceipts(context.Background(), since, until)

	if gotMin == "" {
		t.Error("ListReceipts() did not send created_at_min query param")
	}
	if gotMax == "" {
		t.Error("ListReceipts() did not send created_at_max query param")
	}
}

func TestListReceipts_multiPage(t *testing.T) {
	r1 := receiptFixture()
	r2 := loyverse.Receipt{ReceiptNumber: "R-002", ReceiptType: "SALE", Status: "DONE"}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /receipts", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			mustWriteJSON(t, w, map[string]any{
				"receipts": []loyverse.Receipt{r1},
				"cursor":   "page-2",
			})
		} else {
			mustWriteJSON(t, w, map[string]any{
				"receipts": []loyverse.Receipt{r2},
				"cursor":   "",
			})
		}
	})
	c := newTestClient(t, mux)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	got, err := c.ListReceipts(context.Background(), since, until)
	if err != nil {
		t.Fatalf("ListReceipts() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListReceipts() returned %d receipts, want 2", len(got))
	}
}
