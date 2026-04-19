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

func storeFixture() loyverse.Store {
	ts := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
	return loyverse.Store{
		ID:          "store-1",
		Name:        "Main Store",
		Address:     "123 Main St",
		PhoneNumber: "+1-555-0100",
		Description: "Our flagship location",
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}
}

func TestListStores(t *testing.T) {
	store := storeFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stores", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"stores": []loyverse.Store{store},
		})
	})
	c := newTestClient(t, mux)

	got, err := c.ListStores(context.Background())
	if err != nil {
		t.Fatalf("ListStores() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListStores() returned %d stores, want 1", len(got))
	}
	if diff := cmp.Diff(store, got[0]); diff != "" {
		t.Errorf("ListStores()[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestListStores_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stores", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"stores": []any{}})
	})
	c := newTestClient(t, mux)

	got, err := c.ListStores(context.Background())
	if err != nil {
		t.Fatalf("ListStores() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListStores() returned %d stores, want 0", len(got))
	}
}

func TestListStores_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stores", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	c := newTestClient(t, mux)

	_, err := c.ListStores(context.Background())
	if err == nil {
		t.Fatal("ListStores() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetStore(t *testing.T) {
	store := storeFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stores/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, store)
	})
	c := newTestClient(t, mux)

	got, err := c.GetStore(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("GetStore() error = %v", err)
	}
	if diff := cmp.Diff(store, *got); diff != "" {
		t.Errorf("GetStore() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetStore_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stores/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetStore(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetStore() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
