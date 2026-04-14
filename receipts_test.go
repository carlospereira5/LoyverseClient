package loyverse_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/carlospereira5/loyverse"
)

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
