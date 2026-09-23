package feedback

import (
	"context"
	"errors"
	"testing"
)

// fakeRepository is a minimal in-memory Repository, used to test service.go's
// validation without any real HTTP calls.
type fakeRepository struct {
	listFilter   ListFilter
	listEntries  []FeedbackEntry
	listHasMore  bool
	listNextCur  int64
	listErr      error
	triageID     int64
	triageUpdate TriageUpdate
	triageErr    error
}

func (f *fakeRepository) List(_ context.Context, filter ListFilter) ([]FeedbackEntry, bool, int64, error) {
	f.listFilter = filter
	return f.listEntries, f.listHasMore, f.listNextCur, f.listErr
}

func (f *fakeRepository) UpdateTriage(_ context.Context, id int64, update TriageUpdate) error {
	f.triageID = id
	f.triageUpdate = update
	return f.triageErr
}

func TestService_List_ClampsLimit(t *testing.T) {
	cases := []struct {
		name      string
		inLimit   int
		wantLimit int
	}{
		{"zero uses the default", 0, defaultListLimit},
		{"negative uses the default", -5, defaultListLimit},
		{"over 200 uses the default", 500, defaultListLimit},
		{"in range is passed through unchanged", 25, 25},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &fakeRepository{}
			svc := NewService(repo)

			if _, _, _, err := svc.List(context.Background(), ListFilter{Limit: c.inLimit}); err != nil {
				t.Fatalf("List: %v", err)
			}
			if repo.listFilter.Limit != c.wantLimit {
				t.Errorf("repo received Limit = %d, want %d", repo.listFilter.Limit, c.wantLimit)
			}
		})
	}
}

func TestService_List_PropagatesRepositoryError(t *testing.T) {
	repo := &fakeRepository{listErr: ErrNotConfigured}
	svc := NewService(repo)

	_, _, _, err := svc.List(context.Background(), ListFilter{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestService_UpdateTriage_RejectsInvalidStatus(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)

	err := svc.UpdateTriage(context.Background(), 1, TriageUpdate{Status: "Not A Real Status"})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestService_UpdateTriage_RejectsNonPositiveID(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)

	if err := svc.UpdateTriage(context.Background(), 0, TriageUpdate{Status: "New"}); err == nil {
		t.Fatal("expected an error for id 0, got nil")
	}
}

func TestService_UpdateTriage_NormalizesSignOffBlockerAndTrimsStatus(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)

	if err := svc.UpdateTriage(context.Background(), 7, TriageUpdate{Status: "  Assigned  ", SignOffBlocker: "maybe"}); err != nil {
		t.Fatalf("UpdateTriage: %v", err)
	}
	if repo.triageID != 7 {
		t.Errorf("repo received id = %d, want 7", repo.triageID)
	}
	if repo.triageUpdate.Status != "Assigned" {
		t.Errorf("repo received Status = %q, want trimmed %q", repo.triageUpdate.Status, "Assigned")
	}
	if repo.triageUpdate.SignOffBlocker != "No" {
		t.Errorf("repo received SignOffBlocker = %q, want default %q for an unrecognized value", repo.triageUpdate.SignOffBlocker, "No")
	}
}

func TestService_UpdateTriage_AcceptsYesSignOffBlocker(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)

	if err := svc.UpdateTriage(context.Background(), 7, TriageUpdate{Status: "Resolved", SignOffBlocker: "Yes"}); err != nil {
		t.Fatalf("UpdateTriage: %v", err)
	}
	if repo.triageUpdate.SignOffBlocker != "Yes" {
		t.Errorf("repo received SignOffBlocker = %q, want %q", repo.triageUpdate.SignOffBlocker, "Yes")
	}
}
