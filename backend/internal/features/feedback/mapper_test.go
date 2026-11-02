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

// TestToFeedbackResponse_FiltersSnapshotPassesThrough pins down that Phase
// 7's filters snapshot map is carried through unchanged (nil stays nil, a
// populated map isn't dropped or reshaped).
func TestToFeedbackResponse_FiltersSnapshotPassesThrough(t *testing.T) {
	entry := FeedbackEntry{FiltersSnapshot: map[string]string{"district": "Kampala", "year": "2026"}}
	got := toFeedbackResponse(entry)
	if len(got.FiltersSnapshot) != 2 || got.FiltersSnapshot["district"] != "Kampala" || got.FiltersSnapshot["year"] != "2026" {
		t.Errorf("FiltersSnapshot = %v, want %v", got.FiltersSnapshot, entry.FiltersSnapshot)
	}

	empty := toFeedbackResponse(FeedbackEntry{})
	if empty.FiltersSnapshot != nil {
		t.Errorf("FiltersSnapshot = %v, want nil for an entry with none set", empty.FiltersSnapshot)
	}
}

// TestToFeedbackResponse_HasComponentSnapshotPassesThrough pins down Phase
// 8's boolean flag — a plain bool passthrough, but worth a test since a
// bool's zero value (false) is easy to accidentally leave correct by
// omission rather than by an actual field mapping.
func TestToFeedbackResponse_HasComponentSnapshotPassesThrough(t *testing.T) {
	got := toFeedbackResponse(FeedbackEntry{HasComponentSnapshot: true})
	if !got.HasComponentSnapshot {
		t.Error("HasComponentSnapshot = false, want true")
	}
	got = toFeedbackResponse(FeedbackEntry{HasComponentSnapshot: false})
	if got.HasComponentSnapshot {
		t.Error("HasComponentSnapshot = true, want false")
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
