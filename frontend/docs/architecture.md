# Frontend Architecture

The frontend is a workspace-based modular frontend prepared for single-spa microfrontends. It keeps a shell app for global composition and separates product areas into standalone apps that can build independently.

## Workspace Shape

```text
frontend/
  apps/
    shell/
    announcements/
    audit/
    clients/
    data-visualizer/
    documents/
    e-services/
    email/
    feedback-triage/
    issue-tracker/
    report-browser/
    surveillance/
    users/
    utilities/
  packages/
    api/
    auth/
    config/
    microfrontend/
    state/
    types/
    ui/
    utils/
```

## Shell App

`frontend/apps/shell` owns global application composition:

- layouts
- routing composition
- auth guards
- providers
- app orchestration
- shell-owned pages such as home, settings, auth, and newsfeed

The shell should import app route metadata and lifecycle modules, not app internals.

## Standalone Apps

Each app under `frontend/apps/<app>` owns its feature UI:

- pages
- feature components
- nested app routes
- `src/root.component.tsx`
- `src/routes.tsx`
- `src/single-spa.tsx`
- public exports from `src/index.ts`

Standalone apps can run inside the shell or build as independent package artifacts.

## Shared Packages

Packages under `frontend/packages` hold reusable code:

- `api`: RTK Query/API hooks and endpoint modules
- `auth`: auth bootstrap, auth slice, selectors, and auth types
- `config`: app constants and runtime config helpers
- `microfrontend`: lifecycle, route, and runtime prop contracts
- `state`: global Redux store and shared state slices
- `types`: shared domain/API types
- `ui`: shared components, providers, status/severity helpers
- `utils`: shared utility functions

If code is reused across multiple apps, move it into a package instead of importing another app's internals.

## Import Rules

- Apps may import shared packages through `@moh-sso/*`.
- Shell may import app lifecycle modules such as `@moh-sso/users/single-spa`.
- Shell may import app route metadata through public package exports.
- Apps should not import shell code.
- App-to-app imports must be intentional and declared in `package.json`.
- Shared reusable UI, types, API hooks, and utilities should live in `frontend/packages`.

## Alias Rules

- Use `@moh-sso/*` for public app/package imports.
- Use `@/*` only for shell-local imports.
- Removed compatibility aliases:
  - `@/features`
  - `@/store`
  - `@/shared`
  - `@/lib`
  - `@/ui`
  - `@/utils`

## Dependency Metadata

Every app/package has npm-style package metadata:

- `name`
- `version`
- `license`
- `exports`
- `publishConfig`
- `dependencies`
- `peerDependencies`

Run this after import or manifest changes:

```bash
npm run audit:packages
```
