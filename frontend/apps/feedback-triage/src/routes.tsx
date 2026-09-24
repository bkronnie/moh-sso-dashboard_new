import type { MicrofrontendRoute } from "@moh-sso/microfrontend";

export const feedbackTriageRoute: MicrofrontendRoute = {
  appName: "@moh-sso/feedback-triage",
  path: "/apps/dwh/feedback-triage",
  requiredPermissions: ["feedback:read"],
  requiredSystems: ["data-statistics"],
};
