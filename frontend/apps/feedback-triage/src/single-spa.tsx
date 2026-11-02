import { createReactMicrofrontendLifecycle } from "@moh-sso/microfrontend";

import { FeedbackTriageRoot } from "./root.component";

export const { bootstrap, mount, unmount } = createReactMicrofrontendLifecycle(FeedbackTriageRoot);
