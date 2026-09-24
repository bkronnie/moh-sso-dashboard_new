package feedback

// triageUpdateRequest is the PATCH /feedback/:id/triage request body, from
// this backend's own frontend.
type triageUpdateRequest struct {
	Status         string `json:"status" binding:"required"`
	AssignedTo     string `json:"assigned_to"`
	SignOffBlocker string `json:"sign_off_blocker"`
	JiraProject    string `json:"jira_project"`
	JiraEpic       string `json:"jira_epic"`
	JiraSprint     string `json:"jira_sprint"`
	JiraKey        string `json:"jira_key"`
}

type triageUpdateResponse struct {
	Message string `json:"message"`
}

// FeedbackResponse is this backend's own (snake_case, matching the rest of
// this API's response shapes) wire contract for one feedback entry — see
// mapper.go for the translation from the proxied dwh_visualization shape
// (FeedbackEntry, repository.go).
type FeedbackResponse struct {
	ID       int64  `json:"id"`
	ReportID string `json:"report_id"`

	// SectionID is the stable Section.ID (real only for entries submitted
	// after 2026-09-23 on the dwh_visualization side — see FeedbackEntry's
	// doc comment). SectionTitle is the display label. SectionDisplay is
	// SectionTitle if set, else SectionID — the fallback dwh_visualization's
	// own frontend uses, so old and new rows both render a readable label
	// with no extra logic needed on this side.
	SectionID      string `json:"section_id,omitempty"`
	SectionTitle   string `json:"section_title,omitempty"`
	SectionDisplay string `json:"section_display,omitempty"`
	SubsectionID   string `json:"subsection_id,omitempty"`
	ComponentID    string `json:"component_id,omitempty"`
	ComponentTitle string `json:"component_title,omitempty"`
	Username       string `json:"username"`
	Email          string `json:"email,omitempty"`

	Institution      string `json:"institution,omitempty"`
	ReporterRole     string `json:"reporter_role,omitempty"`
	LevelOfOperation string `json:"level_of_operation,omitempty"`
	Phone            string `json:"phone,omitempty"`

	FeedbackType              string `json:"feedback_type,omitempty"`
	Severity                  string `json:"severity,omitempty"`
	Comment                   string `json:"comment"`
	ExpectedResult            string `json:"expected_result,omitempty"`
	AffectsDecisionMaking     string `json:"affects_decision_making,omitempty"`
	ImpactDescription         string `json:"impact_description,omitempty"`
	FollowUpNeeded            string `json:"follow_up_needed,omitempty"`
	CorrectiveActionCandidate string `json:"corrective_action_candidate,omitempty"`
	Recommendation            string `json:"recommendation,omitempty"`

	Status         string `json:"status"`
	AssignedTo     string `json:"assigned_to,omitempty"`
	SignOffBlocker string `json:"sign_off_blocker,omitempty"`

	JiraProject string `json:"jira_project,omitempty"`
	JiraEpic    string `json:"jira_epic,omitempty"`
	JiraSprint  string `json:"jira_sprint,omitempty"`
	JiraKey     string `json:"jira_key,omitempty"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at,omitempty"`

	// FiltersSnapshot (Phase 7): the reporter's active report filters at
	// submission time, empty for entries submitted before it shipped.
	FiltersSnapshot map[string]string `json:"filters_snapshot,omitempty"`

	// HasComponentSnapshot (Phase 8): whether a frozen component render is
	// stored for this entry on the dwh_visualization side — drives the
	// "View Snapshot" link's enabled state.
	HasComponentSnapshot bool `json:"has_component_snapshot"`
}

// FeedbackListResponse is the GET /feedback response envelope.
type FeedbackListResponse struct {
	Entries    []FeedbackResponse `json:"entries"`
	HasMore    bool               `json:"has_more"`
	NextCursor int64              `json:"next_cursor,omitempty"`
}
