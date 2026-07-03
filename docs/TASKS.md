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

### TASK-002 — Add shared MCP handoff bridge for OpenCode and Caveman

Status: IN PROGRESS

Goal:
Add a local Phase 1 MCP broker so OpenCode and Caveman Code can exchange structured handoff tasks without manual prompt relay.

Acceptance criteria:

- `tools/agent-bridge/server.mjs` exists and implements Phase 1 handoff tools.
- `tools/agent-bridge/package.json` exists with MCP SDK dependency.
- `tmp/agent-bridge/.gitignore` protects runtime queue state.
- `opencode.json` and `.cave/settings.json` register the broker.
- Usage is documented in `docs/AGENT_BRIDGE.md` and the OpenCode handoff prompt is updated.
- Local verification confirms broker syntax and dependency install.

Files likely involved:

- `tools/agent-bridge/server.mjs`
- `tools/agent-bridge/package.json`
- `tmp/agent-bridge/.gitignore`
- `opencode.json`
- `.cave/settings.json`
- `.opencode/prompts/caveman-handoff.md`
- `docs/AGENT_BRIDGE.md`

Checks to run:

- `cd tools/agent-bridge && npm install`
- `node --check tools/agent-bridge/server.mjs`
- broker startup smoke test

Notes / risks:

- OpenCode and Caveman are separate endpoints; live MCP exposure still needs in-session verification.
- `tmp/agent-bridge/` must stay out of commits except for `.gitignore`.
- `tools/agent-bridge/node_modules/` must stay untracked.

### TASK-001 — Initialize project from upstream code and template

Status: IN PROGRESS

Goal:
Initialize this repository with `docker-image-watcher` project code, set up remotes, fill project docs, and verify tooling state.

Acceptance criteria:

- `origin` points to Forgejo project remote.
- `upstream` points to GitHub source repository.
- `opencode.json` stays local-only via `.gitignore`.
- Template files coexist with upstream project code.
- GitNexus index is refreshed for current repository state.
- Project docs are filled with real project content.

Files likely involved:

- `.gitignore`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `docs/ROADMAP.md`
- `docs/STATUS.md`
- `docs/TASKS.md`
- `memory/PROJECT_BRIEF.md`
- `memory/CONSTRAINTS.md`
- template configuration directories restored from backup branch

Checks to run:

- `git remote -v`
- `git log --oneline -5`
- `git status`
- `gitnexus analyze`
- `gitnexus status`
- `go build ./...` if Go toolchain is available

Notes / risks:

- `git reset --hard origin/main` is destructive and requires successful fetch first.
- Upstream project keeps Go application files in repository root, not `src/`.
- Live local config must not be staged or pushed.
