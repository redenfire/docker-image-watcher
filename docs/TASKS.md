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

### TASK-003 — Verify live OpenCode/Caveman tooling sessions on cleaned main

Status: IN PROGRESS

Goal:
Verify that the cleaned private `main` branch works end-to-end for live OpenCode and Caveman sessions, including GitNexus visibility and `agent_bridge` availability where configured.

Acceptance criteria:

- OpenCode session confirms GitNexus MCP visibility.
- Caveman/Cave session confirms GitNexus access through MCP or accepted CLI fallback.
- `agent_bridge` visibility is checked in both endpoints after session restart.
- Any endpoint-specific config mismatch is documented in repo docs.

Current progress:

- `opencode mcp list` now shows both `gitnexus` and `agent_bridge` after local `tools/agent-bridge` dependency install and local-only `opencode.json` alignment.
- `gitnexus status` works from project root, so Caveman CLI fallback is available if MCP tools are not exposed in-session.
- Interactive Caveman-session MCP visibility is still pending manual verification in a fresh session.

Files likely involved:

- `docs/STATUS.md`
- `docs/CAVEMAN_GITNEXUS.md`
- `docs/AGENT_BRIDGE.md`
- `.cave/settings.json`
- `.cave/mcp.json`
- local-only `opencode.json` if user updates local OpenCode config

Checks to run:

- `gitnexus status`
- `opencode mcp list`
- `opencode mcp debug gitnexus`
- in-session Caveman/Cave MCP visibility checks

Notes / risks:

- OpenCode and Caveman are separate endpoints and may load different MCP config surfaces.
- `opencode.json` is local-only and must not be committed.
- Session restarts may be required after config changes.

## Completed tasks

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
