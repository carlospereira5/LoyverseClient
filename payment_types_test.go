package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func TestListPaymentTypes_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /payment_types", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"payment_types": []any{}})
	})
	c := newTestClient(t, mux)

	got, err := c.ListPaymentTypes(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentTypes() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListPaymentTypes() returned %d payment types, want 0", len(got))
	}
}

func TestListPaymentTypes_populated(t *testing.T) {
	pt := paymentTypeFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /payment_types", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"payment_types": []loyverse.PaymentType{pt},
		})
	})
	c := newTestClient(t, mux)

	got, err := c.ListPaymentTypes(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentTypes() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListPaymentTypes() returned %d payment types, want 1", len(got))
	}
	if diff := cmp.Diff(pt, *got[0]); diff != "" {
		t.Errorf("ListPaymentTypes()[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestListPaymentTypes_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /payment_types", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	c := newTestClient(t, mux)

	_, err := c.ListPaymentTypes(context.Background())
	if err == nil {
		t.Fatal("ListPaymentTypes() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetPaymentType(t *testing.T) {
	pt := paymentTypeFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /payment_types/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, pt)
	})
	c := newTestClient(t, mux)

	got, err := c.GetPaymentType(context.Background(), "pt-1")
	if err != nil {
		t.Fatalf("GetPaymentType() error = %v", err)
	}
	if diff := cmp.Diff(pt, *got); diff != "" {
		t.Errorf("GetPaymentType() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetPaymentType_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /payment_types/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetPaymentType(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetPaymentType() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
