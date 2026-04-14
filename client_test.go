package loyverse_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/carlospereira5/loyverse"
)

func TestNew_emptyToken(t *testing.T) {
	_, err := loyverse.New("")
	if err == nil {
		t.Error("New(\"\") = nil error, want error")
	}
}

func TestNew_validToken(t *testing.T) {
	_, err := loyverse.New("my-token")
	if err != nil {
		t.Errorf("New(\"my-token\") error = %v, want nil", err)
	}
}

func TestClient_setsAuthorizationHeader(t *testing.T) {
	var gotHeader string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		mustWriteJSON(t, w, map[string]any{"items": []any{}})
	})
	c := newTestClient(t, mux)

	_, _ = c.GetItems(context.Background())

	if want := "Bearer test-token"; gotHeader != want {
		t.Errorf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestClient_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetItem(context.Background(), "missing")
	if err == nil {
		t.Fatal("GetItem() = nil error, want *loyverse.APIError")
	}

	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}

func TestClient_serverError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	c := newTestClient(t, mux)

	_, err := c.GetItem(context.Background(), "item-1")

	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
}
