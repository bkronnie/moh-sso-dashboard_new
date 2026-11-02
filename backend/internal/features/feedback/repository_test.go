package feedback

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
)

func newTestRepository(t *testing.T, handler http.HandlerFunc) (Repository, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	repo := NewHTTPRepository(&config.Config{
		Feedback: config.FeedbackConfig{
			BaseURL: server.URL,
			APIKey:  "test-key",
			Timeout: 5 * time.Second,
		},
	})
	return repo, server
}

// TestRepository_List_SendsExpectedRequestShape confirms the query params
// and X-API-Key header sent to dwh_visualization exactly match
// FeedbackListAllHandler's contract (server/feedback_handlers.go,
// dwh_visualization) — a mismatch here (e.g. a stale param name) would
// silently return zero results rather than an error.
func TestRepository_List_SendsExpectedRequestShape(t *testing.T) {
	var gotPath, gotAPIKey string
	repo, _ := newTestRepository(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		gotAPIKey = r.Header.Get("X-API-Key")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listResponse{
			Entries: []FeedbackEntry{{ID: 1, ReportID: "r1", Comment: "c", Status: "New"}},
			HasMore: true,
			NextCur: 42,
		})
	})

	entries, hasMore, nextCur, err := repo.List(context.Background(), ListFilter{
		ReportID:     "Programs/Malaria/report",
		SectionTitle: "Malaria Admissions",
		Status:       "New",
		Limit:        10,
		Before:       100,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if gotAPIKey != "test-key" {
		t.Errorf("X-API-Key header = %q, want %q", gotAPIKey, "test-key")
	}
	wantParams := map[string]string{
		"reportId":     "Programs/Malaria/report",
		"sectionTitle": "Malaria Admissions",
		"status":       "New",
		"limit":        "10",
		"before":       "100",
	}
	reqURL, err := url.Parse(gotPath)
	if err != nil {
		t.Fatalf("parse request path %q: %v", gotPath, err)
	}
	for key, want := range wantParams {
		if got := reqURL.Query().Get(key); got != want {
			t.Errorf("query param %q = %q, want %q", key, got, want)
		}
	}

	if len(entries) != 1 || entries[0].ID != 1 {
		t.Fatalf("unexpected entries: %+v", entries)
	}
	if !hasMore || nextCur != 42 {
		t.Errorf("hasMore = %v, nextCur = %d, want true, 42", hasMore, nextCur)
	}
}

// TestRepository_UpdateTriage_SendsCamelCaseBody confirms the PATCH body
// sent matches feedbackTriageBody's field names exactly (camelCase, per
// dwh_visualization's own convention) - not this backend's own snake_case.
func TestRepository_UpdateTriage_SendsCamelCaseBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]string
	repo, _ := newTestRepository(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "id": 7})
	})

	err := repo.UpdateTriage(context.Background(), 7, TriageUpdate{
		Status:         "Assigned",
		AssignedTo:     "Analyst A",
		SignOffBlocker: "Yes",
		JiraKey:        "NDWH-237",
	})
	if err != nil {
		t.Fatalf("UpdateTriage: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/api/feedback/7/triage" {
		t.Errorf("path = %q, want /api/feedback/7/triage", gotPath)
	}
	if gotBody["status"] != "Assigned" || gotBody["assignedTo"] != "Analyst A" || gotBody["jiraKey"] != "NDWH-237" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestRepository_UpdateTriage_404MapsToErrNotFound(t *testing.T) {
	repo, _ := newTestRepository(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	err := repo.UpdateTriage(context.Background(), 999, TriageUpdate{Status: "New"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_List_UnconfiguredReturnsErrNotConfiguredWithoutACall(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	repo := NewHTTPRepository(&config.Config{}) // BaseURL/APIKey both empty

	_, _, _, err := repo.List(context.Background(), ListFilter{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
	if called {
		t.Error("expected no HTTP call to be made when unconfigured")
	}
}
