type RuntimeAppConfig = {
  API_BASE_URL?: string;
  DWH_BASE_URL?: string;
};

function getRuntimeApiBaseUrl() {
  if (typeof globalThis === "undefined" || !("window" in globalThis)) {
    return undefined;
  }

  return (globalThis.window as Window & { __APP_CONFIG__?: RuntimeAppConfig }).__APP_CONFIG__
    ?.API_BASE_URL;
}

function getRuntimeDwhBaseUrl() {
  if (typeof globalThis === "undefined" || !("window" in globalThis)) {
    return undefined;
  }

  return (globalThis.window as Window & { __APP_CONFIG__?: RuntimeAppConfig }).__APP_CONFIG__
    ?.DWH_BASE_URL;
}

const API_BASE_URL =
  getRuntimeApiBaseUrl() ||
  import.meta.env.VITE_API_BASE_URL ||
  "https://dashboards.health.go.ug/ssobackend";

// dwh_visualization's own public site (not its API — that's FEEDBACK_SERVICE_BASE_URL,
// server-to-server only) — used for browser-facing cross-links, e.g. Feedback
// Triage's "Preview" link into Report Catalog's Live Data modal
// (Feedback_Triage_Migration_Work_Plan.md Phase 6).
const DWH_BASE_URL =
  getRuntimeDwhBaseUrl() ||
  import.meta.env.VITE_DWH_BASE_URL ||
  "https://dashboards.health.go.ug";

const API_ROOT = `${API_BASE_URL}/api`;
const API_VERSION = "v1";
const API_BASE = `${API_ROOT}/${API_VERSION}`;

// Use local proxy for visualizer endpoints in development if VITE_VISUALIZER_API_BASE is configured
const VISUALIZER_API_BASE = import.meta.env.VITE_VISUALIZER_API_BASE || API_BASE;

export const API = {
  serviceBase: API_BASE_URL,
  root: API_ROOT,
  version: API_VERSION,
  base: API_BASE,
  build: {
    version: () => `${API_BASE_URL}/version`,
  },

  // --------------------------------------------------
  // Auth
  // --------------------------------------------------
  auth: {
    base: `${API_BASE}/auth`,
    login: () => `${API_BASE}/auth/login`,
    launch: () => `${API_BASE}/auth/launch`,
    callback: () => `${API_BASE}/auth/callback`,
    refresh: () => `${API_BASE}/auth/refresh`,
    logout: () => `${API_BASE}/auth/logout`,
    me: () => `${API_BASE}/auth/me`,
  },

  // --------------------------------------------------
  // Clients
  // --------------------------------------------------
  clients: {
    base: `${API_BASE}/clients`,
    list: () => `${API_BASE}/clients`,
    byId: (id: string) => `${API_BASE}/clients/${id}`,
    create: () => `${API_BASE}/clients`,
    delete: (id: string) => `${API_BASE}/clients/${id}`,
    update: (id: string) => `${API_BASE}/clients/${id}`,

    // -------- Client roles --------
    roles: {
      // READ (authenticated)
      list: (clientId: string) => `${API_BASE}/clients/${clientId}/roles`,

      // WRITE (admin)
      create: (clientId: string) => `${API_BASE}/admin/clients/${clientId}/roles`,
      delete: (clientId: string, role: string) =>
        `${API_BASE}/admin/clients/${clientId}/roles/${encodeURIComponent(role)}`,
    },
  },

  // --------------------------------------------------
  // Users (authenticated)
  // --------------------------------------------------
  users: {
    base: `${API_BASE}/users`,
    list: () => `${API_BASE}/users`,
    byId: (id: string) => `${API_BASE}/users/${id}`,
    create: () => `${API_BASE}/users`,
    delete: (id: string) => `${API_BASE}/users/${id}`,
    update: (id: string) => `${API_BASE}/users/${id}`,
    resetPassword: (id: string) => `${API_BASE}/users/${id}/reset-password`,
  },

  // --------------------------------------------------
  // Visualizer
  // --------------------------------------------------
  visualizer: {
    themes: () => `${VISUALIZER_API_BASE}/visualizer/themes`,
    theme: () => `${VISUALIZER_API_BASE}/visualizer/dataelements/theme`,
    hierarchy: () => `${VISUALIZER_API_BASE}/visualizer/adminunits/hierarchy`,
    dataValues: () => `${VISUALIZER_API_BASE}/visualizer/datavalues`,
    datasets: () => `${VISUALIZER_API_BASE}/visualizer/datasets`,
    dataElements: () => `${VISUALIZER_API_BASE}/visualizer/dataelements`,
  },
  // --------------------------------------------------
  // Visualizer
  // --------------------------------------------------
  issue: {
    list: () => `${VISUALIZER_API_BASE}/issues`,
  },

  // --------------------------------------------------
  // Feedback Triage (proxies dwh_visualization server-to-server)
  // --------------------------------------------------
  feedback: {
    list: () => `${API_BASE}/feedback`,
    triage: (id: number | string) => `${API_BASE}/feedback/${id}/triage`,
  },

  // --------------------------------------------------
  // dwh_visualization (browser-facing cross-links, not API calls)
  // --------------------------------------------------
  dwh: {
    // Opens dwh_visualization's Report Catalog straight into a component's
    // Live Data preview — the exact section/subsection/component a feedback
    // entry was submitted about. componentId is preferred when known (a
    // stable id); omit it and the link still opens, falling back to the
    // report's Mapping modal on the dwh_visualization side (see
    // report-catalog.js's openReportFromURLParam) since old feedback rows
    // (pre-2026-09-23) never captured one. See Feedback_Triage_Migration_Work_Plan.md
    // Phase 6 for the full design. sectionId/subsectionId (Phase 9 tier 3):
    // harmless to always pass when known — dwh_visualization only reads them
    // when componentId doesn't resolve to anything (section/report-level
    // feedback), landing on the first visual in that subsection/section
    // instead of dropping straight to the mapping table.
    reportCatalogPreview: (reportId: string, componentId?: string, sectionId?: string, subsectionId?: string) => {
      const params = new URLSearchParams({ reportId, mode: "live" });
      if (componentId) params.set("componentId", componentId);
      if (subsectionId) params.set("subsectionId", subsectionId);
      if (sectionId) params.set("sectionId", sectionId);
      return `${DWH_BASE_URL}/report-catalog.html?${params.toString()}`;
    },

    // Opens the same Report Catalog Preview modal, but straight into the
    // "As Submitted" tab — a frozen render of the component captured at
    // submission time (Phase 8), not a fresh query. componentId is required
    // here (unlike reportCatalogPreview's graceful fallback) because a
    // snapshot only exists for component-level feedback in the first place —
    // callers should only build this link when entry.has_component_snapshot
    // is true. See Feedback_Triage_Migration_Work_Plan.md Phase 8.
    feedbackSnapshotPreview: (reportId: string, componentId: string, feedbackId: number | string) => {
      const params = new URLSearchParams({ reportId, mode: "snapshot", componentId, feedbackId: String(feedbackId) });
      return `${DWH_BASE_URL}/report-catalog.html?${params.toString()}`;
    },
  },

  // --------------------------------------------------
  // Data validation
  // --------------------------------------------------
  dataValidation: {
    rules: {
      list: () => `${API_BASE}/data-validation/rules`,
      import: () => `${API_BASE}/data-validation/rules/import`,
    },
    dqa: {
      tables: {
        list: () => `${API_BASE}/data-validation/dqa/tables`,
        upsert: () => `${API_BASE}/data-validation/dqa/tables`,
        remove: (tableId: string) => `${API_BASE}/data-validation/dqa/tables/${tableId}`,
      },
      rules: {
        list: (tableId?: string) =>
          tableId
            ? `${API_BASE}/data-validation/dqa/rules?table_id=${encodeURIComponent(tableId)}`
            : `${API_BASE}/data-validation/dqa/rules`,
        upsert: () => `${API_BASE}/data-validation/dqa/rules`,
        compile: () => `${API_BASE}/data-validation/dqa/rules/compile`,
        seed: () => `${API_BASE}/data-validation/dqa/rules/seed`,
        remove: (tableId: string, code: string) =>
          `${API_BASE}/data-validation/dqa/rules/${encodeURIComponent(tableId)}/${encodeURIComponent(code)}`,
      },
      run: () => `${API_BASE}/data-validation/dqa/run`,
      runs: {
        list: (tableId?: string) =>
          tableId
            ? `${API_BASE}/data-validation/dqa/runs?table_id=${encodeURIComponent(tableId)}`
            : `${API_BASE}/data-validation/dqa/runs`,
        flags: (runId: number | string, severity?: string) =>
          severity
            ? `${API_BASE}/data-validation/dqa/runs/${runId}/flags?severity=${encodeURIComponent(severity)}`
            : `${API_BASE}/data-validation/dqa/runs/${runId}/flags`,
      },
    },
  },

  // --------------------------------------------------
  // Admin
  // --------------------------------------------------
  admin: {
    base: `${API_BASE}/admin`,

    // -------- Users --------
    users: {
      list: () => `${API_BASE}/admin/users`,
      byId: (id: string) => `${API_BASE}/admin/users/${id}`,
      create: () => `${API_BASE}/admin/users`,
      update: (id: string) => `${API_BASE}/admin/users/${id}`,
      delete: (id: string) => `${API_BASE}/admin/users/${id}`,
      toggle: (id: string) => `${API_BASE}/admin/users/${id}/toggle`,
      resetPassword: (id: string) => `${API_BASE}/admin/users/${id}/reset-password`,
      passwordResetEmail: (id: string) => `${API_BASE}/admin/users/${id}/password-reset`,

      import: {
        preview: () => `${API_BASE}/admin/users/import/preview`,
        execute: () => `${API_BASE}/admin/users/import/execute`,
        job: (jobId: string) => `${API_BASE}/admin/users/import/${jobId}`,
        errorsCsv: (jobId: string) => `${API_BASE}/admin/users/import/${jobId}/errors.csv`,
        templateCsv: () => `${API_BASE}/admin/users/import/template.csv`,
      },

      // -------- User ↔ Client roles --------
      clientRoles: {
        // GET roles user has for a client
        list: (userId: string, clientId: string) =>
          `${API_BASE}/admin/users/${userId}/clients/${clientId}/roles`,

        // PUT diff-based update
        update: (userId: string) => `${API_BASE}/admin/users/${userId}/client-roles`,

        // POST single role
        assign: (userId: string, clientId: string) =>
          `${API_BASE}/admin/users/${userId}/clients/${clientId}/roles`,

        // DELETE single role
        remove: (userId: string, clientId: string, role: string) =>
          `${API_BASE}/admin/users/${userId}/clients/${clientId}/roles/${role}`,
      },
    },

    // -------- Metrics --------
    metrics: {
      overview: () => `${API_BASE}/admin/metrics/overview`,

      system: {
        countUsers: () => `${API_BASE}/admin/metrics/system/count-users`,
        countDisabledUsers: () => `${API_BASE}/admin/metrics/system/count-disabled-users`,
        activeToday: () => `${API_BASE}/admin/metrics/system/active-today`,
        activeThisWeek: () => `${API_BASE}/admin/metrics/system/active-this-week`,
        loginTrend: () => `${API_BASE}/admin/metrics/system/login-trend`,
        loginTrendRange: () => `${API_BASE}/admin/metrics/system/login-trend-range`,
      },

      security: {
        failedLogins: () => `${API_BASE}/admin/metrics/security/failed-logins`,
        failedLoginsRange: () => `${API_BASE}/admin/metrics/security/failed-logins-range`,
        suspiciousLogins: () => `${API_BASE}/admin/metrics/security/suspicious-logins`,
      },

      clients: {
        count: () => `${API_BASE}/admin/metrics/clients/count`,
        mostAccessed: () => `${API_BASE}/admin/metrics/clients/most-accessed`,
        loginCount: () => `${API_BASE}/admin/metrics/clients/login-count`,
        activeToday: () => `${API_BASE}/admin/metrics/clients/active-today`,
      },

      users: {
        newRange: () => `${API_BASE}/admin/metrics/users/new-range`,
        newTrend: () => `${API_BASE}/admin/metrics/users/new-trend`,
        neverLoggedIn: () => `${API_BASE}/admin/metrics/users/never-logged-in`,
        lastLogin: (userId: string) => `${API_BASE}/admin/metrics/users/last-login/${userId}`,
        clientUsage: (userId: string) => `${API_BASE}/admin/metrics/users/client-usage/${userId}`,
      },
    },

    // -------- Audit --------
    audit: {
      base: `${API_BASE}/admin/audit-logs`,
      list: () => `${API_BASE}/admin/audit-logs`,
      actions: () => `${API_BASE}/admin/audit-logs/actions`,
      byId: (id: string) => `${API_BASE}/admin/audit-logs/${id}`,

      metrics: {
        overview: () => `${API_BASE}/admin/audit-logs/metrics/overview`,
        failedLoginsByDay: () => `${API_BASE}/admin/audit-logs/metrics/failed-logins-by-day`,
        topFailureIps: () => `${API_BASE}/admin/audit-logs/metrics/top-failure-ips`,
      },

      export: () => `${API_BASE}/admin/audit-logs/export`,
    },

    // -------- Notifications --------
    notifications: {
      base: `${API_BASE}/admin/notifications`,
      notify: () => `${API_BASE}/admin/notifications`,
      list: () => `${API_BASE}/admin/notifications`,
      byId: (id: string) => `${API_BASE}/admin/notifications/${id}`,
      markAsRead: (id: string) => `${API_BASE}/admin/notifications/${id}/read`,
      delete: (id: string) => `${API_BASE}/admin/notifications/${id}`,
      count: () => `${API_BASE}/admin/notifications/count`,
      countUnread: () => `${API_BASE}/admin/notifications/count/unread`,
      deleteOld: () => `${API_BASE}/admin/notifications/cleanup`,
      deliveries: (id: string) => `${API_BASE}/admin/notifications/${id}/deliveries`,
      deliveryList: () => `${API_BASE}/admin/notifications/deliveries`,
      deliveryMetrics: () => `${API_BASE}/admin/notifications/deliveries/metrics`,
      deliveryById: (deliveryId: string) =>
        `${API_BASE}/admin/notifications/deliveries/${deliveryId}`,
      retryDelivery: (deliveryId: string) =>
        `${API_BASE}/admin/notifications/deliveries/${deliveryId}/retry`,
      cancelDelivery: (deliveryId: string) =>
        `${API_BASE}/admin/notifications/deliveries/${deliveryId}/cancel`,
      testSms: () => `${API_BASE}/admin/notifications/test-sms`,
    },
  },

  // --------------------------------------------------
  // Notifications (authenticated user)
  // --------------------------------------------------
  notifications: {
    preferences: () => `${API_BASE}/notifications/preferences`,
  },

  // --------------------------------------------------
  // Health (not versioned)
  // --------------------------------------------------
  health: () => `/health`,
} as const;
