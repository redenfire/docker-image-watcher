# Learnings

Durable knowledge discovered while working on the project.

## Format

```md
### YYYY-MM-DD - Short title

What was learned:

Why it matters:

Files/components affected:
```

### 2026-06-23 - Forgejo registry workflows should separate runner push host from client pull host

What was learned:

A Forgejo Actions runner may need to push to a local/LAN registry endpoint that is different from the hostname used later by Portainer or other clients.

Why it matters:

Template workflows and docs should treat `REGISTRY_PUSH_HOST` and `REGISTRY_PULL_HOST` as separate operational concerns so self-hosted CI does not depend on the public pull hostname.

Files/components affected:

- `deploy/freellmapi/builder-repo/.forgejo/workflows/build-freellmapi-image.yaml`
- `deploy/freellmapi/builder-repo/README.md`

### 2026-06-23 - Forgejo repo variables should be sanitized before composing Docker tags or API URLs

What was learned:

Repository variables pasted into Forgejo can carry trailing `\r`/`\n`. If a workflow interpolates them directly into Docker image references or Forgejo API URLs, builds and cleanup calls can fail with invalid references or malformed hosts. Recomputing critical image tags inside each shell step is also easier to debug than depending on fragile image-name propagation from earlier steps.

Why it matters:

Template workflows should strip trailing newlines from registry/image variables before use, and should compose critical Docker tags locally inside each shell step.

Files/components affected:

- `deploy/freellmapi/builder-repo/.forgejo/workflows/build-freellmapi-image.yaml`
- `deploy/freellmapi/builder-repo/.forgejo/workflows/cleanup-freellmapi-registry.yaml`
- `deploy/freellmapi/builder-repo/README.md`
- `deploy/freellmapi/README.md`

### 2026-07-02 - Portable Go toolchain is available locally for build checks

What was learned:

This workstation now has a local portable Go toolchain at `C:\Users\neomod\AppData\Local\PortableTools\Go\go1.26.4`. It is intentionally not installed globally and not added to PATH, so build checks should call `go.exe` by full path when needed.

Why it matters:

Future build verification for this repo can run without changing system-wide tool configuration or assuming `go` is on PATH.

Files/components affected:

- local workstation tool state only

### 2026-07-02 - OpenCode and Caveman interop fits a shared MCP broker better than Markdown relay

What was learned:

OpenCode and Caveman Code are separate endpoints with separate MCP registration paths. Because both can consume project-local MCP servers, the safest path to structured interoperability is a shared local broker with JSON handoffs, not manual Markdown baton-passing.

Why it matters:

Future automation can build on the same broker protocol for launch/orchestration without depending on prompt copy/paste or ad-hoc file parsing.

Files/components affected:

- `tools/agent-bridge/server.mjs`
- `docs/AGENT_BRIDGE.md`
- `opencode.json`
- `.cave/settings.json`

### 2026-07-03 - Private Forgejo main must stay separate from upstream PR branches

What was learned:

This repo needs a private integration branch that carries local tooling/docs, while upstream contribution branches must be cut fresh from `upstream/main`. Reusing private `main` or stale rebased PR branches for upstream submissions can mix local-only scaffold state with upstreamable fixes and create confusing review history.

Why it matters:

Future upstream work should start from latest `upstream/main`, contain only minimal upstreamable patches, and delete stale `pr/*` branches after merges instead of stacking more rebases onto them.

Files/components affected:

- `docs/DECISIONS.md`
- `docs/IMPLEMENTATION_WORKFLOW.md`
- `docs/STATUS.md`
- `AGENTS.md`
- `CLAUDE.md`

### 2026-07-03 - OpenCode broker visibility depends on repo-local agent-bridge dependencies

What was learned:

`opencode mcp list` showed `agent_bridge` as failed until `tools/agent-bridge` dependencies were installed locally. After `cd tools/agent-bridge && npm install`, OpenCode immediately showed both `gitnexus` and `agent_bridge` as connected.

Why it matters:

OpenCode MCP visibility for repo-local brokers is not just a config problem; dependency install state under the tool folder can be the entire cause of a missing or closed MCP connection.

Files/components affected:

- `tools/agent-bridge/package.json`
- `tools/agent-bridge/package-lock.json`
- `docs/AGENT_BRIDGE.md`
- `docs/TASKS.md`
- local-only `opencode.json`

### 2026-07-03 - Cleaned main now has working OpenCode and Caveman tooling path

What was learned:

After upstream sync cleanup, the repo's private `main` branch can support both OpenCode MCP usage and Caveman session usage successfully. GitNexus works from the project root, `agent_bridge` works once local broker dependencies are installed, and the cleaned branch passes Forgejo build.

Why it matters:

Future task work can proceed from cleaned `main` without treating the tooling stack as a blocker. Remaining gaps are optimization/documentation issues, not baseline operability.

Files/components affected:

- `docs/STATUS.md`
- `docs/TASKS.md`
- `memory/LEARNINGS.md`

### 2026-07-04 - Git author email must follow remote destination

What was learned:

This repo needs different author-email behavior depending on where commit history is going. Forgejo `origin` uses the default local email, while GitHub-targeted history (`gh-fork`, and `upstream` if explicitly used) must be created with `-c user.email="n3omod@gmail.com"` on commit-producing commands.

Why it matters:

Pushes do not change author metadata after the fact. The correct email must be applied when the commit is created or rewritten so GitHub-visible attribution is correct without polluting Forgejo-only history.

Files/components affected:

- `.opencode/prompts/caveman-handoff.md`
- `docs/CONTRIBUTING.md`
- `memory/CONSTRAINTS.md`
- `memory/LEARNINGS.md`

### 2026-07-04 - os.Kill cannot be trapped; replacing SIGTERM with it breaks graceful shutdown

What was learned:

When replacing deprecated signal constants, `os.Kill` is not a valid replacement for `syscall.SIGTERM`. `os.Kill` maps to `SIGKILL`, which the process cannot trap, intercept, or handle. Using it in `signal.Notify` would terminate the application immediately without running graceful shutdown logic.

The correct approach is to keep `syscall.SIGTERM` and replace only `SIGINT` with `os.Interrupt`.

Files/components affected:

- `main.go: shutdown signal handling`
- `memory/LEARNINGS.md`

### 2026-07-04 - Registry digest failures must not be treated as update availability

What was learned:

When remote digest lookup fails because of Docker Hub rate limiting or similar registry errors, marking containers as `outdated` is misleading and can trigger unnecessary auto-update attempts. Those cases should surface as `unknown` until a later successful registry check resolves the remote digest.

Why it matters:

The UI and auto-update loop should only offer updates when the application has a confirmed remote digest mismatch, not when registry access is degraded.

Files/components affected:

- `main.go: checkAll`
- `docker.go: pullImageStream, IsRateLimitError`
- `web/index.html`
- `memory/LEARNINGS.md`
