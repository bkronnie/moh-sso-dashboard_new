import { useEffect, useMemo, useState } from "react";
import { Button, Dropdown, InlineNotification, Loading, TextInput } from "@carbon/react";
import { Download, Launch } from "@carbon/react/icons";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import { API } from "@moh-sso/config";
import { useToast } from "@moh-sso/ui";

import DataList from "../../../data-visualizer/src/pages/components/data-table/data-table.component.tsx";
import { useLazyGetFeedbackQuery, useUpdateFeedbackTriageMutation } from "../api";
import { FEEDBACK_STATUSES, type FeedbackEntry, type GetFeedbackParams } from "../types";

import "./feedback-triage.scss";

// Fixed enums, mirrored from dwh_visualization's own submission form
// (public/index.html, #feedback-type / #feedback-severity options) — not
// re-derived from a live call, since these are a stable, small, shared
// contract between the submission form and this triage view.
const FEEDBACK_TYPES = [
  "Data issue",
  "Calculation issue",
  "Interpretation / clarity issue",
  "Display issue",
  "Other",
];
const SEVERITIES = ["Low", "Medium", "High"];
const SIGN_OFF_OPTIONS = ["No", "Yes"];

const PAGE_LIMIT = 50;

// Shape DataList/Carbon's DataTable expect: {key, header} per column, and
// each row object keyed the same way. Row objects below also carry the
// full FeedbackEntry (spread in), since DataList hands the matched row
// object straight to renderExpandedContent — that's how the detail/edit
// form gets the fields (comment, jira*, etc.) this column set doesn't show.
//
// status/severity render as plain text, not colored Carbon Tag elements —
// tried once (2026-09-24), reverted the same day: the shared DataList
// component (data-table.component.tsx) forces a table remount by computing
// `JSON.stringify(paginatedData)` on the full row set. That's safe for plain
// strings/numbers, but a row field holding a real React element (a <Tag>)
// picks up a circular `_owner`/Fiber reference once React actually renders
// it client-side, and JSON.stringify throws on that — which, with no error
// boundary in this app, unmounts the whole page (confirmed live: header
// rendered, then blanked once real feedback rows loaded and got stringified).
// An `renderToStaticMarkup` (SSR) test beforehand missed this because SSR
// doesn't build a Fiber tree at all, so it never exercises the circular
// reference — SSR-testing a client-rendering concern was the actual mistake,
// not the Tag idea itself. Re-adding colored badges needs the shared
// component's remount-key made safe against non-serializable cell values
// first (e.g. a JSON.stringify replacer that special-cases React elements),
// not just reverted here.
const columnHeaders = [
  { key: "report", header: "Report" },
  { key: "section", header: "Section" },
  { key: "type", header: "Type" },
  { key: "severity", header: "Severity" },
  { key: "status", header: "Status" },
  { key: "assignedTo", header: "Assigned To" },
  { key: "createdAt", header: "Submitted" },
];

function formatDate(unixSeconds?: number) {
  if (!unixSeconds) return "-";
  return new Date(unixSeconds * 1000).toLocaleString();
}

// Phase 7: renders the reporter's active report filters at submission time
// as "district=Kampala, year=2026" — empty/absent for entries submitted
// before this shipped. Mirrors dwh_visualization's own
// formatFiltersSnapshot (public/js/feedback-triage.js).
function formatFiltersSnapshot(filters?: Record<string, string>): string {
  if (!filters) return "";
  return Object.entries(filters)
    .filter(([, v]) => v)
    .map(([k, v]) => `${k}=${v}`)
    .join(", ");
}

// Program Area isn't a field dwh_visualization's feedback rows carry
// directly — it's derived client-side from report_id's path (e.g.
// "Programs/Malaria/malaria-national-monthly-report" -> "Programs/Malaria"),
// from whatever's already loaded. Avoids a second call to dwh_visualization
// just to populate this one dropdown (the plan's explicit simplification
// over the original dwh_visualization page, which called GET /api/reports
// for the same purpose).
function deriveProgramArea(reportId: string): string {
  const segments = reportId.split("/").filter(Boolean);
  return segments.length > 1 ? segments.slice(0, -1).join("/") : reportId;
}

function toCsv(rows: FeedbackEntry[]): string {
  const cols: Array<[string, (e: FeedbackEntry) => string]> = [
    ["ID", (e) => String(e.id)],
    ["Report", (e) => e.report_id],
    ["Section", (e) => e.section_display ?? ""],
    ["Type", (e) => e.feedback_type ?? ""],
    ["Severity", (e) => e.severity ?? ""],
    ["Status", (e) => e.status],
    ["Assigned To", (e) => e.assigned_to ?? ""],
    ["Sign-off Blocker", (e) => e.sign_off_blocker ?? ""],
    ["Jira Project", (e) => e.jira_project ?? ""],
    ["Jira Epic", (e) => e.jira_epic ?? ""],
    ["Jira Sprint", (e) => e.jira_sprint ?? ""],
    ["Jira Key", (e) => e.jira_key ?? ""],
    ["Comment", (e) => e.comment],
    ["Submitted By", (e) => e.username],
    ["Submitted At", (e) => formatDate(e.created_at)],
    ["Filters at Submission", (e) => formatFiltersSnapshot(e.filters_snapshot)],
  ];
  const escape = (value: string) => `"${value.replace(/"/g, '""')}"`;
  const header = cols.map(([label]) => escape(label)).join(",");
  const lines = rows.map((row) => cols.map(([, get]) => escape(get(row))).join(","));
  return [header, ...lines].join("\n");
}

function downloadCsv(csv: string) {
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `feedback-triage-${new Date().toISOString().slice(0, 10)}.csv`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

type DraftFilters = {
  status: string;
  programArea: string;
  severity: string;
  type: string;
  reportId: string;
  sectionTitle: string;
  assignedTo: string;
};

const emptyFilters: DraftFilters = {
  status: "",
  programArea: "",
  severity: "",
  type: "",
  reportId: "",
  sectionTitle: "",
  assignedTo: "",
};

function toQueryParams(filters: DraftFilters): GetFeedbackParams {
  return {
    status: filters.status || undefined,
    severity: filters.severity || undefined,
    type: filters.type || undefined,
    report_id: filters.reportId || undefined,
    section_title: filters.sectionTitle || undefined,
    assigned_to: filters.assignedTo || undefined,
    limit: PAGE_LIMIT,
  };
}

export default function FeedbackTriage() {
  const toast = useToast();

  const [draft, setDraft] = useState<DraftFilters>(emptyFilters);
  const [appliedFilters, setAppliedFilters] = useState<DraftFilters>(emptyFilters);
  // Program Area filters client-side over whatever's currently loaded
  // (deriveProgramArea), rather than being sent to the server — dwh_visualization
  // has no such field to filter on directly.
  const [programAreaFilter, setProgramAreaFilter] = useState("");

  const [fetchFeedback, { isFetching, error }] = useLazyGetFeedbackQuery();
  const [entries, setEntries] = useState<FeedbackEntry[]>([]);
  const [nextCursor, setNextCursor] = useState<number | undefined>();
  const [hasMore, setHasMore] = useState(false);

  const runQuery = async (filters: DraftFilters, before: number | undefined, append: boolean) => {
    const params = { ...toQueryParams(filters), before };
    const result = await fetchFeedback(params).unwrap();
    setEntries((prev) => (append ? [...prev, ...result.entries] : result.entries));
    setHasMore(result.has_more);
    setNextCursor(result.next_cursor);
  };

  useEffect(() => {
    runQuery(appliedFilters, undefined, false).catch(() => {
      // surfaced via `error` from the hook; nothing else to do here
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [appliedFilters]);

  const applyFilters = () => setAppliedFilters(draft);
  const clearFilters = () => {
    setDraft(emptyFilters);
    setAppliedFilters(emptyFilters);
    setProgramAreaFilter("");
  };
  const loadMore = () => runQuery(appliedFilters, nextCursor, true);

  const programAreas = useMemo(() => {
    const set = new Set(entries.map((e) => deriveProgramArea(e.report_id)));
    return Array.from(set).sort();
  }, [entries]);

  const visibleEntries = useMemo(() => {
    if (!programAreaFilter) return entries;
    return entries.filter((e) => deriveProgramArea(e.report_id) === programAreaFilter);
  }, [entries, programAreaFilter]);

  // DataList/Carbon's DataTable need a row object per entry keyed by the
  // column `key`s above, plus a stable string `id`. The full entry is
  // spread in too, since DataList passes this same object through to
  // renderExpandedContent — that's where the detail/edit form reads
  // comment/jira*/etc., fields this column set doesn't display.
  const tableRows = useMemo(
    () =>
      visibleEntries.map((entry) => ({
        ...entry,
        id: String(entry.id),
        report: entry.report_id,
        section: entry.section_display ?? "-",
        type: entry.feedback_type ?? "-",
        severity: entry.severity ?? "-",
        status: entry.status,
        assignedTo: entry.assigned_to ?? "-",
        createdAt: formatDate(entry.created_at),
      })),
    [visibleEntries],
  );

  const handleSaved = (updated: FeedbackEntry) => {
    setEntries((prev) => prev.map((e) => (e.id === updated.id ? updated : e)));
    toast.success("Triage fields updated");
  };
  const handleSaveError = () => toast.error("Failed to update triage fields");

  return (
    <div className="feedback-triage-page">
      <div className="feedback-triage-header">
        <h1>Feedback Triage</h1>
        <p>Feedback submitted from dwh_visualization report pages.</p>
      </div>

      <div className="feedback-triage-filters">
        <Dropdown
          id="filter-status"
          titleText="Status"
          label="All"
          items={["", ...FEEDBACK_STATUSES]}
          itemToString={(item) => (item ? item : "All")}
          selectedItem={draft.status}
          onChange={({ selectedItem }) => setDraft((d) => ({ ...d, status: selectedItem ?? "" }))}
        />
        <Dropdown
          id="filter-program-area"
          titleText="Program Area"
          label="All"
          items={["", ...programAreas]}
          itemToString={(item) => (item ? item : "All")}
          selectedItem={programAreaFilter}
          onChange={({ selectedItem }) => setProgramAreaFilter(selectedItem ?? "")}
        />
        <Dropdown
          id="filter-severity"
          titleText="Severity"
          label="All"
          items={["", ...SEVERITIES]}
          itemToString={(item) => (item ? item : "All")}
          selectedItem={draft.severity}
          onChange={({ selectedItem }) => setDraft((d) => ({ ...d, severity: selectedItem ?? "" }))}
        />
        <Dropdown
          id="filter-type"
          titleText="Type"
          label="All"
          items={["", ...FEEDBACK_TYPES]}
          itemToString={(item) => (item ? item : "All")}
          selectedItem={draft.type}
          onChange={({ selectedItem }) => setDraft((d) => ({ ...d, type: selectedItem ?? "" }))}
        />
        <TextInput
          id="filter-report"
          labelText="Report"
          placeholder="Report id"
          value={draft.reportId}
          onChange={(e) => setDraft((d) => ({ ...d, reportId: e.target.value }))}
        />
        <TextInput
          id="filter-section"
          labelText="Section"
          placeholder="Section title"
          value={draft.sectionTitle}
          onChange={(e) => setDraft((d) => ({ ...d, sectionTitle: e.target.value }))}
        />
        <TextInput
          id="filter-assigned-to"
          labelText="Assigned To"
          value={draft.assignedTo}
          onChange={(e) => setDraft((d) => ({ ...d, assignedTo: e.target.value }))}
        />

        <div className="feedback-triage-filter-actions">
          <Button kind="primary" size="md" onClick={applyFilters}>
            Apply
          </Button>
          <Button kind="ghost" size="md" onClick={clearFilters}>
            Clear
          </Button>
          <Button
            kind="tertiary"
            size="md"
            renderIcon={Download}
            disabled={visibleEntries.length === 0}
            onClick={() => downloadCsv(toCsv(visibleEntries))}
          >
            Export CSV
          </Button>
        </div>
      </div>

      {error ? (
        <InlineNotification
          kind="error"
          title="Failed to load feedback"
          subtitle="The feedback service may not be configured yet, or dwh_visualization is unreachable."
          lowContrast
        />
      ) : null}

      <div className="feedback-triage-table-container">
        {isFetching && visibleEntries.length === 0 ? (
          <div className="feedback-triage-loading-state">
            <Loading small withOverlay={false} />
          </div>
        ) : (
          <>
            <DataList
              columns={columnHeaders}
              data={tableRows}
              renderExpandedContent={(row) => (
                <FeedbackDetailForm entry={row as FeedbackEntry} onSaved={handleSaved} onSaveError={handleSaveError} />
              )}
            />
            {/* A filter-triggered refetch keeps the previous rows on screen
                (isFetching true, visibleEntries still non-empty) — a small
                inline spinner rather than replacing the table avoids the
                whole list flickering away and back on every Apply click. */}
            {isFetching ? <Loading small withOverlay={false} /> : null}
          </>
        )}

        {!isFetching && visibleEntries.length === 0 ? (
          <div className="feedback-triage-empty">
            <p>No feedback entries match the current filters.</p>
          </div>
        ) : null}
      </div>

      {hasMore && !programAreaFilter ? (
        <div className="feedback-triage-load-more">
          <Button kind="secondary" onClick={loadMore} disabled={isFetching}>
            Load more
          </Button>
        </div>
      ) : null}
    </div>
  );
}

type FeedbackDetailFormProps = {
  entry: FeedbackEntry;
  onSaved: (entry: FeedbackEntry) => void;
  onSaveError: () => void;
};

function FeedbackDetailForm({ entry, onSaved, onSaveError }: FeedbackDetailFormProps) {
  const [updateTriage, { isLoading: isSaving }] = useUpdateFeedbackTriageMutation();

  const [status, setStatus] = useState(entry.status);
  const [assignedTo, setAssignedTo] = useState(entry.assigned_to ?? "");
  const [signOffBlocker, setSignOffBlocker] = useState(entry.sign_off_blocker || "No");
  const [jiraProject, setJiraProject] = useState(entry.jira_project ?? "");
  const [jiraEpic, setJiraEpic] = useState(entry.jira_epic ?? "");
  const [jiraSprint, setJiraSprint] = useState(entry.jira_sprint ?? "");
  const [jiraKey, setJiraKey] = useState(entry.jira_key ?? "");

  const save = async () => {
    try {
      await updateTriage({
        id: entry.id,
        body: {
          status,
          assigned_to: assignedTo,
          sign_off_blocker: signOffBlocker,
          jira_project: jiraProject,
          jira_epic: jiraEpic,
          jira_sprint: jiraSprint,
          jira_key: jiraKey,
        },
      }).unwrap();
      onSaved({
        ...entry,
        status,
        assigned_to: assignedTo,
        sign_off_blocker: signOffBlocker,
        jira_project: jiraProject,
        jira_epic: jiraEpic,
        jira_sprint: jiraSprint,
        jira_key: jiraKey,
      });
    } catch {
      onSaveError();
    }
  };

  // Links out to dwh_visualization's Report Catalog, straight into a Live
  // Data preview of the exact component this feedback was submitted about --
  // see Feedback_Triage_Migration_Work_Plan.md Phase 6. Deliberately a plain
  // link to another site's page, not a re-render of the component here: this
  // app doesn't duplicate dwh_visualization's chart/table rendering logic.
  // Always clickable, even with no component_id (rows submitted before
  // 2026-09-23, or subsection-only feedback) -- dwh_visualization's own
  // openReportFromURLParam() already falls back gracefully to the report's
  // Mapping modal when it can't resolve one specific component, so this
  // link is never a dead end, just sometimes less precise.
  //
  // "View Snapshot" (Phase 8) answers a different question -- "what did the
  // reporter actually see," a frozen render captured at submission time, vs.
  // Preview's "does this look like this today." Unlike Preview, genuinely
  // disabled when there's nothing to show: no snapshot was captured (old
  // entry, section/report-level feedback, or capture failed), so there's no
  // graceful fallback to offer.

  return (
    <div className="feedback-triage-detail">
      <div className="feedback-triage-detail-comment">
        <div className="feedback-triage-detail-comment-header">
          <strong>Comment ({entry.username}):</strong>
          <div className="feedback-triage-detail-comment-links">
            <Button
              kind="ghost"
              size="sm"
              renderIcon={Launch}
              href={API.dwh.reportCatalogPreview(entry.report_id, entry.component_id, entry.section_id, entry.subsection_id)}
              target="_blank"
              rel="noopener"
            >
              Preview in Report Catalog
            </Button>
            {entry.has_component_snapshot && entry.component_id ? (
              <Button
                kind="ghost"
                size="sm"
                renderIcon={Launch}
                href={API.dwh.feedbackSnapshotPreview(entry.report_id, entry.component_id, entry.id)}
                target="_blank"
                rel="noopener"
              >
                View Snapshot
              </Button>
            ) : (
              <Button
                kind="ghost"
                size="sm"
                renderIcon={Launch}
                disabled
                title="No snapshot was captured for this entry (it may predate this feature, or have no single component to snapshot)"
              >
                View Snapshot
              </Button>
            )}
          </div>
        </div>
        <p>{entry.comment}</p>
        {entry.recommendation ? (
          <>
            <strong>Recommendation:</strong>
            <p>{entry.recommendation}</p>
          </>
        ) : null}
        {formatFiltersSnapshot(entry.filters_snapshot) ? (
          <p className="feedback-triage-filters-snapshot">
            <strong>Filters at submission:</strong> {formatFiltersSnapshot(entry.filters_snapshot)}
          </p>
        ) : null}
      </div>

      <PermissionGuard
        permission={PERMISSIONS.feedbackTriage}
        fallback={<p className="feedback-triage-readonly-note">You don't have permission to update triage fields.</p>}
      >
        <div className="feedback-triage-form">
          <Dropdown
            id={`status-${entry.id}`}
            titleText="Status"
            label={status}
            items={[...FEEDBACK_STATUSES]}
            selectedItem={status}
            onChange={({ selectedItem }) => setStatus(selectedItem ?? status)}
          />
          <TextInput
            id={`assigned-to-${entry.id}`}
            labelText="Assigned To"
            value={assignedTo}
            onChange={(e) => setAssignedTo(e.target.value)}
          />
          <Dropdown
            id={`sign-off-${entry.id}`}
            titleText="Sign-off Blocker"
            label={signOffBlocker}
            items={SIGN_OFF_OPTIONS}
            selectedItem={signOffBlocker}
            onChange={({ selectedItem }) => setSignOffBlocker(selectedItem ?? signOffBlocker)}
          />
          <TextInput
            id={`jira-project-${entry.id}`}
            labelText="Jira Project"
            value={jiraProject}
            onChange={(e) => setJiraProject(e.target.value)}
          />
          <TextInput
            id={`jira-epic-${entry.id}`}
            labelText="Jira Epic"
            value={jiraEpic}
            onChange={(e) => setJiraEpic(e.target.value)}
          />
          <TextInput
            id={`jira-sprint-${entry.id}`}
            labelText="Jira Sprint"
            value={jiraSprint}
            onChange={(e) => setJiraSprint(e.target.value)}
          />
          <TextInput
            id={`jira-key-${entry.id}`}
            labelText="Jira Key"
            value={jiraKey}
            onChange={(e) => setJiraKey(e.target.value)}
          />
          <Button kind="primary" onClick={save} disabled={isSaving}>
            {isSaving ? "Saving..." : "Save"}
          </Button>
        </div>
      </PermissionGuard>
    </div>
  );
}
