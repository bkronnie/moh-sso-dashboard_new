# Feedback Triage App

The feedback triage app lets the data team review, assign, and track feedback that users submit from dwh_visualization report pages.

It is a standalone React microfrontend exposed as `@moh-sso/feedback-triage` and mounted by the shell at `/apps/dwh/feedback-triage`.

## Responsibilities

- List feedback entries with filters (status, program area, severity, type, report, section, assignee) and CSV export.
- Show each entry's comment, recommendation, and the report filters active when it was submitted.
- Link out to dwh_visualization's Report Catalog for a live preview, and to the frozen snapshot when one was captured.
- Update triage fields (status, assignee, sign-off blocker, Jira project/epic/sprint/key) for users with `feedback:triage`.
- Expose single-spa lifecycle functions for shell orchestration.

## Backend

This app calls the portal backend, not dwh_visualization directly:

- `GET /api/v1/feedback` (needs `feedback:read`)
- `PATCH /api/v1/feedback/:id/triage` (needs `feedback:triage`)

The backend (`backend/internal/features/feedback`) proxies these to dwh_visualization with an X-API-Key. Set `FEEDBACK_SERVICE_BASE_URL` and `FEEDBACK_SERVICE_API_KEY` in `backend/app.env`; with either empty the endpoints return 503. Keycloak roles: `feedback_viewer` (read) and `feedback_manager` (read + triage).

## Entry Points

- `src/index.ts`
- `src/routes.tsx`
- `src/single-spa.tsx`

## Development

```bash
npm run dev:mf:feedback-triage   # dev server on port 4116
npm run build -w @moh-sso/feedback-triage
```

Opening `http://localhost:4116/` directly loads a standalone page (`src/dev-standalone.tsx`) with a fake signed-in admin. It calls `http://localhost:9000`, so run the backend (or a mock of it) there. This page is never part of the library build.
