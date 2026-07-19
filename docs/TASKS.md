# Tasks

This file is the task board for this project.

It is not for:

- installing workstation tools;
- configuring Forgejo;
- maintaining this template;
- recording private assistant scratchpad notes.

Setup steps belong in `docs/HOW_TO_USE.md`.

## Task rules

Agents must work on one task at a time.

Each task should include:

- status;
- goal;
- acceptance criteria;
- files likely involved;
- checks to run;
- notes/risks.

## Active tasks

### TASK-009 — Fix ghcr.io digest matching and batch update flow

Status: IN PROGRESS

Goal:
Fix upstream-divergence issues that misclassify some ghcr.io-backed images and can break grouped update runs after container recreation.

Files likely involved:
- `docker.go`
- `main.go`
- `docs/TASKS.md`
- `docs/STATUS.md`
- `memory/LEARNINGS.md`

Checks to run:
- `go build ./...`
- `go vet ./...`
- `gofmt -w main.go docker.go` produces no semantic changes beyond formatting
- verify ghcr.io-backed groups no longer show false outdated state when another repo digest already matches remote digest
- verify grouped update continues across recreated containers with refreshed `app.images` state

Notes / risks:
- Temporary debug logging was added in `checkAll` to capture raw Docker list image values during live verification.
- Live Docker verification is still required to confirm exact ghcr.io image naming behavior.

### TASK-008 — UI improvements for rate-limit banner, update buttons, and status badges

Status: IN PROGRESS

Goal:
Improve the web UI by fixing rate-limit banner dismissal, reducing redundant per-group update actions, adding a global update button, and improving status badge readability.

Files likely involved:
- `web/index.html`
- `docs/TASKS.md`
- `docs/STATUS.md`

Checks to run:
- visual banner-dismiss verification
- verify single-container groups hide redundant group update button
- verify global update button appears only when outdated containers exist
- verify larger status badges and updated Italian label render correctly

Notes / risks:
- Global update action runs sequential group updates from the browser and still needs live UI verification.

### TASK-007 — Fix auth bugs for email-style usernames, login throttling, and expired-session redirect

Status: IN PROGRESS

Goal:
Fix three auth issues: email-style usernames containing `@`, ineffective login throttling caused by `RemoteAddr` port variance, and expired sessions that leave the UI on a broken page instead of redirecting to login.

Files likely involved:
- `main.go`
- `web/index.html`
- `docs/TASKS.md`
- `docs/STATUS.md`

Checks to run:
- `go build ./...`
- `go vet ./...`
- manual login with `AUTH_USER=user@domain.com`
- verify repeated failed logins trigger temporary block
- verify expired session redirects to `/login.html`

Notes / risks:
- Browser-side expired-session behavior still needs live verification after implementation.
- Login throttle remains process-local in-memory state.

## Completed tasks

### TASK-006 — Multi-registry auth, persistent error display, docs

Status: COMPLETED

Goal:
Extend authenticated pull support to multiple registries, persist last pull errors in the UI, and update matching documentation.

Outcome:
- `docker.go` supports `DOCKER_REGISTRY_AUTH` as either Docker Hub shorthand `username:password` or a JSON registry-to-credentials map.
- `main.go` persists last pull errors per container and exposes them through `GET /api/images`.
- `web/index.html` renders persistent container pull errors in red under each update cell.
- README and project docs now describe multi-registry auth formats, persistent pull errors, and updated troubleshooting guidance.
- `go build ./...` and `go vet ./...` passed after implementation.

### TASK-005 — Add UI warning for Docker pull rate limit + Docker auth upgrade

Status: COMPLETED

Goal:
Add a persistent UI warning when Docker pull rate limiting is detected and support authenticated Docker Engine pulls through environment configuration.

Outcome:
- `docker.go` detects Docker pull rate-limit errors and supports `DOCKER_REGISTRY_AUTH=username:password` for `X-Registry-Auth` pull requests.
- `main.go` tracks current rate-limit state and exposes `GET /api/ratelimit` for the web UI.
- `web/index.html` shows a dismissible warning banner with English and Italian copy when the backend reports rate limiting.
- Project docs were updated for configuration, API, troubleshooting, and current status.
- `go build ./...` and `go vet ./...` passed after implementation.

### TASK-004 — Apply all TIER 1/2/3 fixes as upstream PRs

Status: COMPLETED

Goal:
Identify and fix all code quality issues across Go source files, Dockerfile, and CI workflow. Send each logical fix as a separate upstream PR branch cut from `upstream/main`.

Outcome:
- 12 upstream PR branches created and pushed to `gh-fork`:
  - TIER 1 (4): getImageDigest loop, recreateContainer errors, unmarshal errors, saveAuto errors
  - TIER 2 (4): go fmt, `.dockerignore`, Dockerfile COPY glob, Names index guard
  - TIER 3 (4): io.ReadAll errors, syscall deprecation, fs.Sub error, sort results
- 2 Forgejo-only maintenance commits added on `main`:
  - README multi-arch claim
  - CI cache-from
- All PR branches merged into `main` for Forgejo build
- All builds and vets passed on each branch and on merged `main`
- Forgejo image build succeeded

### TASK-003 — Verify live OpenCode/Caveman tooling sessions on cleaned main

Status: COMPLETED

Goal:
Verify that the cleaned private `main` branch works end-to-end for live OpenCode and Caveman sessions, including GitNexus visibility and `agent_bridge` availability where configured.

Outcome:

- OpenCode session confirmed GitNexus MCP visibility.
- OpenCode session confirmed `agent_bridge` visibility after local broker dependency install.
- Caveman/Cave session verification was completed successfully; tooling is considered operational on cleaned `main`.
- `gitnexus status` works from project root as an available fallback path.
- Endpoint-specific config notes were documented in repo docs.

Notes / risks:

- `opencode.json` remains local-only and must not be committed.
- Repo-local MCP brokers still depend on installed dependencies under `tools/agent-bridge/`.


### TASK-002 — Add shared MCP handoff bridge for OpenCode and Caveman

Status: COMPLETED

Goal:
Add a local Phase 1 MCP broker so OpenCode and Caveman Code can exchange structured handoff tasks without manual prompt relay.

Outcome:

- `tools/agent-bridge/server.mjs` exists and implements Phase 1 handoff tools.
- `tools/agent-bridge/package.json` exists with MCP SDK dependency.
- `tmp/agent-bridge/.gitignore` protects runtime queue state.
- Repo-side Cave config includes broker registration.
- Usage is documented in `docs/AGENT_BRIDGE.md` and the OpenCode handoff prompt is updated.
- Local syntax / startup verification was completed; live endpoint verification moved to TASK-003.

### TASK-001 — Initialize project from upstream code and template

Status: COMPLETED

Goal:
Initialize this repository with `docker-image-watcher` project code, set up remotes, fill project docs, and verify tooling state.

Outcome:

- `origin` points to Forgejo project remote.
- `upstream` points to GitHub source repository.
- `main` was cleaned and resynced onto latest `upstream/main` with private scaffold restored in a separate follow-up commit.
- `opencode.json` stays local-only via `.gitignore`.
- Template/tooling files coexist with upstream project code.
- GitNexus index is refreshed for current repository state.
- Forgejo build passed on cleaned branch state.

Notes / risks:

- Future upstream contribution branches must be cut from fresh `upstream/main`, not private `main`.
- Live local config must remain untracked.
