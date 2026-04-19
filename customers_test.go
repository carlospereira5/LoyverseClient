package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func TestListCustomers_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"customers": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	got, err := c.ListCustomers(context.Background())
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListCustomers() returned %d customers, want 0", len(got))
	}
}

func TestListCustomers_singlePage(t *testing.T) {
	cust := customerFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"customers": []loyverse.Customer{cust},
			"cursor":    "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.ListCustomers(context.Background())
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListCustomers() returned %d customers, want 1", len(got))
	}
	if diff := cmp.Diff(cust, got[0]); diff != "" {
		t.Errorf("ListCustomers()[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestListCustomers_multiPage(t *testing.T) {
	cust1 := customerFixture()
	cust2 := loyverse.Customer{ID: "cust-2", Name: "Bob Builder"}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			mustWriteJSON(t, w, map[string]any{
				"customers": []loyverse.Customer{cust1},
				"cursor":    "page-2",
			})
		} else {
			mustWriteJSON(t, w, map[string]any{
				"customers": []loyverse.Customer{cust2},
				"cursor":    "",
			})
		}
	})
	c := newTestClient(t, mux)

	got, err := c.ListCustomers(context.Background())
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListCustomers() returned %d customers, want 2", len(got))
	}
}

func TestListCustomers_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	c := newTestClient(t, mux)

	_, err := c.ListCustomers(context.Background())
	if err == nil {
		t.Fatal("ListCustomers() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetCustomer(t *testing.T) {
	cust := customerFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, cust)
	})
	c := newTestClient(t, mux)

	got, err := c.GetCustomer(context.Background(), "cust-1")
	if err != nil {
		t.Fatalf("GetCustomer() error = %v", err)
	}
	if diff := cmp.Diff(cust, *got); diff != "" {
		t.Errorf("GetCustomer() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetCustomer_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetCustomer(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetCustomer() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}

func TestCreateOrUpdateCustomer(t *testing.T) {
	want := customerFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /customers", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, want)
	})
	c := newTestClient(t, mux)

	got, err := c.CreateOrUpdateCustomer(context.Background(), loyverse.CustomerRequest{
		Name:  want.Name,
		Email: want.Email,
	})
	if err != nil {
		t.Fatalf("CreateOrUpdateCustomer() error = %v", err)
	}
	if diff := cmp.Diff(want, *got); diff != "" {
		t.Errorf("CreateOrUpdateCustomer() mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateOrUpdateCustomer_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /customers", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"validation error"}`, http.StatusUnprocessableEntity)
	})
	c := newTestClient(t, mux)

	_, err := c.CreateOrUpdateCustomer(context.Background(), loyverse.CustomerRequest{Name: "X"})
	if err == nil {
		t.Fatal("CreateOrUpdateCustomer() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestDeleteCustomer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	c := newTestClient(t, mux)

	if err := c.DeleteCustomer(context.Background(), "cust-1"); err != nil {
		t.Fatalf("DeleteCustomer() error = %v", err)
	}
}

func TestDeleteCustomer_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	err := c.DeleteCustomer(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("DeleteCustomer() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
