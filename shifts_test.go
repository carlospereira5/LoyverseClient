package loyverse_test

import (
	"context"
	"net/http"
	"testing"
	"time"

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
	if got[0].ID != shift.ID {
		t.Errorf("ListShifts()[0].ID = %q, want %q", got[0].ID, shift.ID)
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
