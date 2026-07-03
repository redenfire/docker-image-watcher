# OpenCode Local Instructions

Read `docs/TOOLING_MODEL.md` before assuming which editor, agent, provider, or model route is responsible for a task.


Use the root `AGENTS.md` as the controlling policy.

OpenCode must use GitNexus MCP as the repository memory layer once it is verified.


Default route:

```text
model = deepseek/deepseek-chat
small_model = freellmapi/auto
```

TAB cycles agents, not providers. Use `/models` to switch model route during a session.

## Implementation handoff

The preferred flow is OpenCode planning first, then Caveman Code implementation. Planning agents must produce a scoped implementation prompt for Caveman Code and must not edit files.

This repo intentionally keeps upstream application files at the repository root (`main.go`, `docker.go`, `registry.go`, `web/`, `Dockerfile`, `docker-compose.yml`). Do not invent a `src/` migration or scatter generated source files elsewhere; follow `docs/ARCHITECTURE.md` and existing repo layout.
