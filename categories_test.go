package loyverse_test

import (
	"context"
	"net/http"
	"testing"

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
