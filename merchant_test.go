package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func TestGetMerchant(t *testing.T) {
	merchant := merchantFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /merchant", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, merchant)
	})
	c := newTestClient(t, mux)

	got, err := c.GetMerchant(context.Background())
	if err != nil {
		t.Fatalf("GetMerchant() error = %v", err)
	}
	if diff := cmp.Diff(merchant, *got); diff != "" {
		t.Errorf("GetMerchant() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetMerchant_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /merchant", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	c := newTestClient(t, mux)

	_, err := c.GetMerchant(context.Background())
	if err == nil {
		t.Fatal("GetMerchant() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}
