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

func TestListShifts(t *testing.T) {
	shift := shiftFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /shifts", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, map[string]any{
			"shifts": []loyverse.Shift{shift},
			"cursor": "",
		})
	})
	c := newTestClient(t, mux)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	got, err := c.ListShifts(context.Background(), since, until)
	if err != nil {
		t.Fatalf("ListShifts() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListShifts() returned %d shifts, want 1", len(got))
	}
	if diff := cmp.Diff(shift, got[0]); diff != "" {
		t.Errorf("ListShifts()[0] mismatch (-want +got):\n%s", diff)
	}
}

func TestListShifts_passesDateRangeParams(t *testing.T) {
	var gotMin, gotMax string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /shifts", func(w http.ResponseWriter, r *http.Request) {
		gotMin = r.URL.Query().Get("created_at_min")
		gotMax = r.URL.Query().Get("created_at_max")
		mustWriteJSON(t, w, map[string]any{"shifts": []any{}, "cursor": ""})
	})
	c := newTestClient(t, mux)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)
	_, _ = c.ListShifts(context.Background(), since, until)

	if gotMin == "" {
		t.Error("ListShifts() did not send created_at_min query param")
	}
	if gotMax == "" {
		t.Error("ListShifts() did not send created_at_max query param")
	}
}

func TestGetShift(t *testing.T) {
	fixture := shiftFixture()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /shifts/{id}", func(w http.ResponseWriter, r *http.Request) {
		mustWriteJSON(t, w, fixture)
	})
	c := newTestClient(t, mux)

	got, err := c.GetShift(context.Background(), "shift-1")
	if err != nil {
		t.Fatalf("GetShift() error = %v", err)
	}
	if diff := cmp.Diff(fixture, *got); diff != "" {
		t.Errorf("GetShift() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetShift_apiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /shifts/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	})
	c := newTestClient(t, mux)

	_, err := c.GetShift(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetShift() = nil error, want *loyverse.APIError")
	}
	var apiErr *loyverse.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("GetShift() error type = %T, want *loyverse.APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}
