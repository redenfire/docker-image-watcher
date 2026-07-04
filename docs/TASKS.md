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

No active tasks currently recorded.

## Completed tasks

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
