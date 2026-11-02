// Mirrors this backend's own snake_case feedback DTOs
// (backend/internal/features/feedback/dto.go), which in turn proxy
// dwh_visualization's feedback API. See
// Feedback_Triage_Migration_Work_Plan.md (dwh_visualization repo) for the
// full design.

export const FEEDBACK_STATUSES = [
  "New",
  "Assigned",
  "Correction in progress",
  "Revalidation pending",
  "Resolved",
  "Closed",
] as const;

export type FeedbackStatus = (typeof FEEDBACK_STATUSES)[number];

export type FeedbackEntry = {
  id: number;
  report_id: string;

  // section_display is section_title if set, else section_id — the same
  // fallback dwh_visualization's own frontend uses, so old (pre
  // 2026-09-23) and new feedback rows both render a readable label. Prefer
  // this over section_id/section_title directly when just displaying a row.
  section_id?: string;
  section_title?: string;
  section_display?: string;
  subsection_id?: string;
  component_id?: string;
  component_title?: string;

  username: string;
  email?: string;

  institution?: string;
  reporter_role?: string;
  level_of_operation?: string;
  phone?: string;

  feedback_type?: string;
  severity?: string;
  comment: string;
  expected_result?: string;
  affects_decision_making?: string;
  impact_description?: string;
  follow_up_needed?: string;
  corrective_action_candidate?: string;
  recommendation?: string;

  status: FeedbackStatus | string;
  assigned_to?: string;
  sign_off_blocker?: string;

  jira_project?: string;
  jira_epic?: string;
  jira_sprint?: string;
  jira_key?: string;

  created_at: number;
  updated_at?: number;

  // The reporter's active report filters (district/year/month/etc.) at
  // submission time. Empty/absent for entries submitted before this
  // shipped (2026-09-24, migration Phase 7).
  filters_snapshot?: Record<string, string>;

  // Whether dwh_visualization has a frozen component render stored for
  // this entry (migration Phase 8) — drives the "View Snapshot" link's
  // enabled state. False for entries with no component_id, entries
  // submitted before this shipped, or where capture failed.
  has_component_snapshot: boolean;
};

export type GetFeedbackParams = {
  report_id?: string;
  program_area?: string;
  section_title?: string;
  status?: string;
  severity?: string;
  type?: string;
  assigned_to?: string;
  limit?: number;
  before?: number;
};

export type GetFeedbackResponse = {
  entries: FeedbackEntry[];
  has_more: boolean;
  next_cursor?: number;
};

export type TriageUpdatePayload = {
  status: string;
  assigned_to?: string;
  sign_off_blocker?: string;
  jira_project?: string;
  jira_epic?: string;
  jira_sprint?: string;
  jira_key?: string;
};
