package feedback

import "testing"

// TestToFeedbackResponse_SectionDisplayFallback pins down the
// SectionTitle-else-SectionID fallback mapper.go computes — the same
// bridge dwh_visualization's own frontend uses so pre- and
// post-2026-09-23 feedback rows both render a readable section label. See
// FeedbackEntry's doc comment (repository.go) for why SectionID alone
// isn't always a real id.
func TestToFeedbackResponse_SectionDisplayFallback(t *testing.T) {
	cases := []struct {
		name    string
		entry   FeedbackEntry
		wantSet string
	}{
		{
			name:    "new row: title set, id real - display is the title",
			entry:   FeedbackEntry{SectionID: "section-88586821", SectionTitle: "Malaria Admissions"},
			wantSet: "Malaria Admissions",
		},
		{
			name:    "old row: title empty, id holds the display text - display falls back to it",
			entry:   FeedbackEntry{SectionID: "Malaria Admissions", SectionTitle: ""},
			wantSet: "Malaria Admissions",
		},
		{
			name:    "neither set - display is empty, not a placeholder",
			entry:   FeedbackEntry{},
			wantSet: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := toFeedbackResponse(c.entry)
			if got.SectionDisplay != c.wantSet {
				t.Errorf("SectionDisplay = %q, want %q", got.SectionDisplay, c.wantSet)
			}
			if got.SectionID != c.entry.SectionID {
				t.Errorf("SectionID = %q, want %q (raw id must still pass through unchanged)", got.SectionID, c.entry.SectionID)
			}
			if got.SectionTitle != c.entry.SectionTitle {
				t.Errorf("SectionTitle = %q, want %q (raw title must still pass through unchanged)", got.SectionTitle, c.entry.SectionTitle)
			}
		})
	}
}

func TestToFeedbackResponses_EmptyInputReturnsEmptySliceNotNil(t *testing.T) {
	got := toFeedbackResponses(nil)
	if got == nil {
		t.Fatal("expected a non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}
