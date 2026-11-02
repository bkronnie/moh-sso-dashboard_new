// Package feedback proxies dwh_visualization's Feedback Triage data —
// dwh_visualization stays the system of record (submission + storage);
// this feature only reads and updates triage fields via its HTTP API.
// See dwh_visualization's local_scripts_docs/work_plans/
// Feedback_Triage_Migration_Work_Plan.md for the full design and the
// architecture decision (backend-to-backend proxy, not a data migration).
package feedback

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/moh-sso-dashboard/internal/config"
)

var (
	// ErrNotConfigured is returned by every Repository method when
	// FEEDBACK_SERVICE_BASE_URL / FEEDBACK_SERVICE_API_KEY aren't set —
	// config.validateFeedbackConfig already rejects a half-configured
	// state, so this only fires when both are genuinely empty.
	ErrNotConfigured = errors.New("feedback proxy is not configured (FEEDBACK_SERVICE_BASE_URL / FEEDBACK_SERVICE_API_KEY)")
	ErrNotFound      = errors.New("feedback entry not found")
)

// FeedbackEntry mirrors dwh_visualization's server/feedback_db.go
// FeedbackEntry wire shape exactly (camelCase JSON, that repo's own
// convention) — this is the proxied domain type. dto.go's FeedbackResponse
// is this backend's own (snake_case) API contract; mapper.go translates
// between the two.
//
// SectionID is the real, stable Section.ID for entries submitted after
// 2026-09-23; older rows still carry the section's display title there
// instead (a since-fixed mislabeling on the dwh_visualization side) —
// SectionTitle is empty for those old rows. See mapper.go's SectionDisplay
// for the fallback this backend's frontend should actually render.
type FeedbackEntry struct {
	ID             int64  `json:"id"`
	ReportID       string `json:"reportId"`
	SectionID      string `json:"sectionId,omitempty"`
	SectionTitle   string `json:"sectionTitle,omitempty"`
	SubsectionID   string `json:"subsectionId,omitempty"`
	ComponentID    string `json:"componentId,omitempty"`
	ComponentTitle string `json:"componentTitle,omitempty"`
	Username       string `json:"username"`
	Email          string `json:"email,omitempty"`

	Institution      string `json:"institution,omitempty"`
	ReporterRole     string `json:"reporterRole,omitempty"`
	LevelOfOperation string `json:"levelOfOperation,omitempty"`
	Phone            string `json:"phone,omitempty"`

	FeedbackType              string `json:"feedbackType,omitempty"`
	Severity                  string `json:"severity,omitempty"`
	Comment                   string `json:"comment"`
	ExpectedResult            string `json:"expectedResult,omitempty"`
	AffectsDecisionMaking     string `json:"affectsDecisionMaking,omitempty"`
	ImpactDescription         string `json:"impactDescription,omitempty"`
	FollowUpNeeded            string `json:"followUpNeeded,omitempty"`
	CorrectiveActionCandidate string `json:"correctiveActionCandidate,omitempty"`
	Recommendation            string `json:"recommendation,omitempty"`

	Status         string `json:"status"`
	AssignedTo     string `json:"assignedTo,omitempty"`
	SignOffBlocker string `json:"signOffBlocker,omitempty"`

	JiraProject string `json:"jiraProject,omitempty"`
	JiraEpic    string `json:"jiraEpic,omitempty"`
	JiraSprint  string `json:"jiraSprint,omitempty"`
	JiraKey     string `json:"jiraKey,omitempty"`

	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt,omitempty"`

	// FiltersSnapshot (added 2026-09-24, migration Phase 7): the reporter's
	// active report filters (district/year/month/etc.) at submission time.
	// Empty for entries submitted before this shipped on the dwh_visualization
	// side.
	FiltersSnapshot map[string]string `json:"filtersSnapshot,omitempty"`

	// HasComponentSnapshot (added 2026-09-24, migration Phase 8): whether
	// dwh_visualization has a frozen component render stored for this entry
	// — the actual snapshot is never sent here, only this boolean (drives
	// the "View Snapshot" link's enabled state); the real payload is fetched
	// by the browser directly from dwh_visualization's report-catalog.html,
	// the same link-out pattern Phase 6's Preview link already uses.
	HasComponentSnapshot bool `json:"hasComponentSnapshot"`
}

// ListFilter is the set of filters GET /api/feedback (dwh_visualization)
// accepts — see FeedbackListAllHandler, server/feedback_handlers.go.
type ListFilter struct {
	ReportID     string
	ProgramArea  string
	SectionTitle string
	Status       string
	Severity     string
	FeedbackType string
	AssignedTo   string
	Limit        int
	Before       int64
}

// listResponse mirrors feedbackListResponse (server/feedback_handlers.go).
type listResponse struct {
	Entries []FeedbackEntry `json:"entries"`
	HasMore bool            `json:"hasMore"`
	NextCur int64           `json:"nextCursor,omitempty"`
}

// TriageUpdate mirrors feedbackTriageBody (server/feedback_handlers.go) —
// the PATCH .../triage payload.
type TriageUpdate struct {
	Status         string
	AssignedTo     string
	SignOffBlocker string
	JiraProject    string
	JiraEpic       string
	JiraSprint     string
	JiraKey        string
}

// Repository talks to dwh_visualization's feedback API. There is
// deliberately no Postgres-backed implementation — the SMS-provider
// adapter shape (backend/internal/service/sms_service.go) was the model
// for keeping this swappable, in case a future local-storage
// implementation ever replaces it, but no such replacement is planned
// (dwh_visualization stays the system of record).
type Repository interface {
	List(ctx context.Context, filter ListFilter) (entries []FeedbackEntry, hasMore bool, nextCursor int64, err error)
	UpdateTriage(ctx context.Context, id int64, update TriageUpdate) error
}

type httpRepository struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewHTTPRepository(cfg *config.Config) Repository {
	return &httpRepository{
		baseURL: cfg.Feedback.BaseURL,
		apiKey:  cfg.Feedback.APIKey,
		client:  &http.Client{Timeout: cfg.Feedback.Timeout},
	}
}

func (r *httpRepository) configured() bool {
	return r != nil && r.baseURL != "" && r.apiKey != ""
}

func (r *httpRepository) List(ctx context.Context, filter ListFilter) ([]FeedbackEntry, bool, int64, error) {
	if !r.configured() {
		return nil, false, 0, ErrNotConfigured
	}

	q := url.Values{}
	if filter.ReportID != "" {
		q.Set("reportId", filter.ReportID)
	}
	if filter.ProgramArea != "" {
		q.Set("programArea", filter.ProgramArea)
	}
	if filter.SectionTitle != "" {
		q.Set("sectionTitle", filter.SectionTitle)
	}
	if filter.Status != "" {
		q.Set("status", filter.Status)
	}
	if filter.Severity != "" {
		q.Set("severity", filter.Severity)
	}
	if filter.FeedbackType != "" {
		q.Set("type", filter.FeedbackType)
	}
	if filter.AssignedTo != "" {
		q.Set("assignedTo", filter.AssignedTo)
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}
	if filter.Before > 0 {
		q.Set("before", strconv.FormatInt(filter.Before, 10))
	}

	var resp listResponse
	if err := r.do(ctx, http.MethodGet, "/api/feedback?"+q.Encode(), nil, &resp); err != nil {
		return nil, false, 0, err
	}
	return resp.Entries, resp.HasMore, resp.NextCur, nil
}

func (r *httpRepository) UpdateTriage(ctx context.Context, id int64, update TriageUpdate) error {
	if !r.configured() {
		return ErrNotConfigured
	}

	body, err := json.Marshal(map[string]string{
		"status":         update.Status,
		"assignedTo":     update.AssignedTo,
		"signOffBlocker": update.SignOffBlocker,
		"jiraProject":    update.JiraProject,
		"jiraEpic":       update.JiraEpic,
		"jiraSprint":     update.JiraSprint,
		"jiraKey":        update.JiraKey,
	})
	if err != nil {
		return fmt.Errorf("marshal triage update: %w", err)
	}

	path := fmt.Sprintf("/api/feedback/%d/triage", id)
	return r.do(ctx, http.MethodPatch, path, bytes.NewReader(body), nil)
}

// do sends one request to dwh_visualization, authenticated via the
// X-API-Key service credential (dwh_visualization's feedbackConditionalAuth,
// server/feedback_handlers.go). A 404 maps to ErrNotFound; any other
// non-2xx becomes an error carrying the upstream response body for
// diagnosability.
func (r *httpRepository) do(ctx context.Context, method, path string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(r.baseURL, "/")+path, body)
	if err != nil {
		return fmt.Errorf("build feedback request: %w", err)
	}
	req.Header.Set("X-API-Key", r.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("call dwh_visualization feedback api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read feedback response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("dwh_visualization feedback api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode feedback response: %w", err)
	}
	return nil
}
