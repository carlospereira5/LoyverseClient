package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func TestListEmployees_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"employees": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	got, err := c.ListEmployees(context.Background())
	if err != nil {
		t.Fatalf("ListEmployees() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListEmployees() returned %d employees, want 0", len(got))
	}
}

func TestListEmployees_singlePage(t *testing.T) {
	emp := employeeFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"employees": []loyverse.Employee{emp},
			"cursor":    "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.ListEmployees(context.Background())
	if err != nil {
		t.Fatalf("ListEmployees() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListEmployees() returned %d employees, want 1", len(got))
	}
	if diff := cmp.Diff(emp, *got[0]); diff != "" {
		t.Errorf("ListEmployees()[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestListEmployees_multiPage(t *testing.T) {
	emp1 := employeeFixture()
	emp2 := loyverse.Employee{ID: "emp-2", Name: "Second Employee", Stores: []string{}}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			mustWriteJSON(t, w, map[string]any{
				"employees": []loyverse.Employee{emp1},
				"cursor":    "page-2",
			})
		} else {
			mustWriteJSON(t, w, map[string]any{
				"employees": []loyverse.Employee{emp2},
				"cursor":    "",
			})
		}
	})
	c := newTestClient(t, mux)

	got, err := c.ListEmployees(context.Background())
	if err != nil {
		t.Fatalf("ListEmployees() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ListEmployees() returned %d employees, want 2", len(got))
	}
}

func TestListEmployees_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	})
	c := newTestClient(t, mux)

	_, err := c.ListEmployees(context.Background())
	if err == nil {
		t.Fatal("ListEmployees() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetEmployee(t *testing.T) {
	emp := employeeFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, emp)
	})
	c := newTestClient(t, mux)

	got, err := c.GetEmployee(context.Background(), "emp-1")
	if err != nil {
		t.Fatalf("GetEmployee() error = %v", err)
	}
	if diff := cmp.Diff(emp, *got); diff != "" {
		t.Errorf("GetEmployee() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetEmployee_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetEmployee(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetEmployee() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
