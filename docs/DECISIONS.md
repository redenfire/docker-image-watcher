# Decisions

This file records durable architectural/project decisions.

Use it when a choice affects future implementation, operations, cost, security, or maintenance.

## Format

### DECISION-001 — Title

Date: YYYY-MM-DD

Status: proposed / accepted / superseded

Context:
TBD.

Decision:
TBD.

Consequences:
TBD.

## Decisions

### DECISION-001 — Preserve upstream root layout during initial import

Date: 2026-07-02

Status: accepted

Context:
Template defaults to putting implementation under `src/`, but upstream `redenfire/docker-image-watcher` keeps `main.go`, `docker.go`, `registry.go`, `Dockerfile`, `docker-compose.yml`, and `web/` at repository root. One roadmap goal is contributing changes back upstream.

Decision:
Keep upstream application layout intact for initial import and future near-term work. Store agentic template material alongside it in dedicated documentation and tooling folders rather than moving Go application files into `src/`.

Consequences:
- Local repository stays closer to upstream for easier diffs and future merge requests.
- Root-level application files are an intentional exception to template default layout.
- Future structural refactors should justify added divergence from upstream.

### DECISION-002 — Use shared local MCP broker for OpenCode and Caveman handoff

Date: 2026-07-02

Status: accepted

Context:
OpenCode and Caveman Code are separate agent endpoints. File-based handoff works as fallback, but it still depends on the user to relay prompts and results manually. Both endpoints already support project-local MCP registration, so a shared local task bus can provide structured interoperability without requiring direct native RPC between the tools.

Decision:
Add a local `agent-bridge` MCP server under `tools/agent-bridge/` and register it in both `opencode.json` and `.cave/settings.json`. Use JSON handoffs stored under `tmp/agent-bridge/` for queued, claimed, completed, and failed work items.

Consequences:
- OpenCode and Caveman can exchange structured task state through shared tools rather than ad-hoc Markdown.
- The broker becomes the preferred handoff path, while `scripts/run-caveman.ps1` remains a fallback.
- Future automation can build on the broker for launch/orchestration without redesigning the handoff schema.

### DECISION-003 — Keep private Forgejo `main` separate from upstream PR branches

Date: 2026-07-03

Status: accepted

Context:
This repository mixes upstream application code with private agentic tooling, local workflow docs, and Forgejo-first project scaffolding that is useful for local work but not appropriate for upstream pull requests. A rebase attempt showed that stacking upstreamable fixes and private scaffold changes on the same branch makes PR history confusing and can accidentally push local-only material into shared history.

Decision:
Use `main` as the private Forgejo integration branch: latest `upstream/main` plus this repository's accepted private tooling/documentation scaffold. Create upstream contribution branches only from fresh `upstream/main`, not from private `main`. Keep local-only files such as `.env`, `opencode.json`, runtime handoff archives, and scratch state untracked and out of any `pr/*` branch.

Consequences:
- Private workflow files can evolve on `main` without polluting upstream contribution branches.
- Every new upstream PR must be cut or re-cut from latest `upstream/main` and contain only upstreamable changes.
- Stale `pr/*` branches should be deleted after upstream merges or superseding rebases instead of being reused blindly.
