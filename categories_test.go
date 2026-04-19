package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/carlospereira5/loyverse"
)

func TestListCategories(t *testing.T) {
	cat := categoryFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /categories", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"categories": []loyverse.Category{cat},
			"cursor":     "",
		})
	})
	c := newTestClient(t, mux)

	got, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListCategories() returned %d categories, want 1", len(got))
	}
	if got[0].ID != cat.ID {
		t.Errorf("ListCategories()[0].ID = %q, want %q", got[0].ID, cat.ID)
	}
	if got[0].Name != cat.Name {
		t.Errorf("ListCategories()[0].Name = %q, want %q", got[0].Name, cat.Name)
	}
}

func TestListCategories_empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /categories", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{"categories": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	got, err := c.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListCategories() returned %d categories, want 0", len(got))
	}
}

func TestCreateOrUpdateCategory(t *testing.T) {
	want := categoryFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /categories", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, want)
	})
	c := newTestClient(t, mux)

	got, err := c.CreateOrUpdateCategory(context.Background(), loyverse.CategoryRequest{Name: want.Name, Color: want.Color})
	if err != nil {
		t.Fatalf("CreateOrUpdateCategory() error = %v", err)
	}
	if diff := cmp.Diff(want, *got); diff != "" {
		t.Errorf("CreateOrUpdateCategory() mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateOrUpdateCategory_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /categories", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"validation error"}`, http.StatusUnprocessableEntity)
	})
	c := newTestClient(t, mux)

	_, err := c.CreateOrUpdateCategory(context.Background(), loyverse.CategoryRequest{Name: "X"})
	if err == nil {
		t.Fatal("CreateOrUpdateCategory() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestDeleteCategory(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	c := newTestClient(t, mux)

	if err := c.DeleteCategory(context.Background(), "cat-1"); err != nil {
		t.Fatalf("DeleteCategory() error = %v", err)
	}
}

func TestDeleteCategory_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	err := c.DeleteCategory(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("DeleteCategory() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
