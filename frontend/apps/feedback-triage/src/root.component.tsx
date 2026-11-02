import type { MicrofrontendRuntimeProps } from "@moh-sso/microfrontend";

import FeedbackTriage from "./pages/feedback-triage.component";

export function FeedbackTriageRoot(props: MicrofrontendRuntimeProps) {
  void props;

  return (
    <div className="moh-microfrontend-root">
      <FeedbackTriage />
    </div>
  );
}
