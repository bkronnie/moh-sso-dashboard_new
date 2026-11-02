// Standalone dev harness -- NOT part of the shell/single-spa build.
//
// Mounts FeedbackTriage in isolation, outside the portal shell, so it can be
// click-tested against real dwh_visualization data without the Dockerized
// moh-sso-dashboard backend running (2026-09-24 decision, after that
// backend turned out to have a pre-existing, unrelated startup crash --
// scope of that fix was kept out of "add feedback triage").
//
// Seeds a fake logged-in admin directly into Redux (real auth/me is never
// called) and points RTK Query at a small local mock proxy
// (local_scripts_docs/scripts/feedback-mock-proxy.mjs in dwh_visualization,
// listening on :9000 -- the same port the real backend would use in dev,
// per public/config.development.js) that forwards to dwh_visualization's
// real /api/feedback endpoints and translates the response shape the same
// way backend/internal/features/feedback/mapper.go does.
//
// Only reachable by visiting this app's own dev server directly
// (http://localhost:4116/) -- never bundled into the microfrontend library
// build (vite build only emits the entries vite.microfrontend.config.ts
// lists: index/routes/single-spa/api/types -- this file isn't one of them),
// so it never reaches the shell or any production artifact.

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Provider } from "react-redux";
import { configureStore } from "@reduxjs/toolkit";

import authReducer, { loginSuccess } from "../../../packages/auth/src/auth/auth.slice";
import { baseApi } from "../../../packages/api/src/api/baseApi";
import { ToastProvider } from "../../../packages/ui/src/notifications";
// The real shell's own header (MoH coat of arms + "MOH Integrated Health
// Portal" branding, packages/ui/src/shell/UserHeader.component.tsx) --
// reused as-is rather than hand-rolled, so this harness looks like the real
// portal instead of a bare page. onNavigateHome/onLogout are no-ops: this
// harness has no routing or real session to act on.
import { UserHeader } from "../../../packages/ui/src/shell";

import FeedbackTriage from "./pages/feedback-triage.component";

import "../../shell/src/index.scss";

const store = configureStore({
  reducer: {
    auth: authReducer,
    [baseApi.reducerPath]: baseApi.reducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(baseApi.middleware),
});

// Full permission set, not just feedback:* -- PermissionGuard components
// elsewhere in the app tree (if any get added later) shouldn't need this
// file updated too.
store.dispatch(
  loginSuccess({
    accessToken: null,
    user: {
      id: "standalone-dev-user",
      username: "dev.tester",
      email: "dev.tester@localhost",
      firstName: "Dev",
      lastName: "Tester",
      fullName: "Dev Tester",
      isAdmin: true,
      isUser: true,
      realmRoles: ["admin", "user"],
      clientRoles: {},
      permissions: ["feedback:read", "feedback:triage"],
      systems: ["data-statistics"],
      accessibleSystems: [],
      enabled: true,
      emailVerified: true,
      requirePwdChange: false,
    },
  }),
);

const container = document.getElementById("root");
if (!container) {
  throw new Error("dev-standalone.tsx: #root element not found in index.html");
}

createRoot(container).render(
  <StrictMode>
    <Provider store={store}>
      <ToastProvider>
        <UserHeader username="dev.tester" onNavigateHome={() => {}} onLogout={() => {}} />
        <div style={{ maxWidth: 1400, margin: "0 auto" }}>
          <FeedbackTriage />
        </div>
      </ToastProvider>
    </Provider>
  </StrictMode>,
);
