# Agent Instructions

Read `docs/TOOLING_MODEL.md` before assuming which editor, agent, provider, or model route is responsible for a task.


This repository is a Forgejo-first upstream-tracking project with private agentic tooling scaffold.

Agents must follow the project documents and avoid free-running implementation. Work must be scoped, verifiable, and tied to `docs/TASKS.md`.

## Canonical project flow

```text
Create Project Folder -> Set Forgejo -> Deploy/verify FreeLLMAPI -> Set up free models -> VS Code -> OpenCode -> load .env -> GitNexus -> verify GitNexus MCP in OpenCode -> Caveman Code -> verify Cave GitNexus access -> paid/frontier routes -> fill project docs -> OpenCode plan -> Caveman Code implementation
```


## Required context before planning or coding

Read these files first:

- `docs/HOW_TO_USE.md`
- `docs/STATUS.md`
- `docs/TASKS.md`
- `docs/ROADMAP.md`
- `docs/ARCHITECTURE.md`
- `memory/PROJECT_BRIEF.md`
- `memory/CONSTRAINTS.md`
- `memory/LEARNINGS.md`


## Implementation handoff

Default serious-work flow:

```text
OpenCode plan first -> user reviews/approves -> Caveman Code implements approved plan
```

OpenCode must analyze the project and produce implementation instructions. Caveman Code executes those instructions as the focused coding endpoint and should use GitNexus through Cave-side MCP configuration or an explicitly reported CLI fallback.

Do not start coding until OpenCode has produced a scoped plan for the active task, unless the user explicitly bypasses the handoff.

## Repository layout rule

This repository intentionally keeps upstream application files at the repository root to stay close to `redenfire/docker-image-watcher`.

Primary application files remain:

```text
main.go
docker.go
registry.go
web/
Dockerfile
docker-compose.yml
```

Do not scatter new root-level files beyond the established upstream app layout plus project documentation, tool configuration, environment examples, scripts, Git/Forgejo metadata, and top-level build files that the project genuinely requires. New auxiliary code should follow `docs/ARCHITECTURE.md` and existing folder conventions instead of inventing a default `src/` migration.

## GitNexus requirement

GitNexus is the required repository memory/index layer.

Before broad file exploration, use GitNexus MCP when available. OpenCode uses `opencode.json`; Caveman Code/Cave uses `.cave/settings.json`. If GitNexus MCP is not available in the active endpoint, stop and report the missing memory/index layer unless the user explicitly accepts GitNexus CLI fallback.

Minimum checks:

```text
gitnexus status
opencode mcp debug gitnexus
# Caveman Code/Cave: verify GitNexus MCP tools or run gitnexus status from the project root
```

## Model routing policy

Default OpenCode route:

```text
model = deepseek/deepseek-chat
small_model = freellmapi/auto
```

Use:

- FreeLLMAPI for low-risk documentation cleanup, summaries, formatting, and cheap/light work.
- DeepSeek for normal OpenCode planning and code implementation.
- OpenAI API only when explicitly selected and intentionally funded.
- Caveman Code for low-token terminal coding and optional ChatGPT/Codex auth path.
- Codex in VS Code only if the user chooses that IDE endpoint.

Never assume ChatGPT/Codex subscription usage is the same as OpenAI API-key billing.

## Work rules

- Work on one task at a time.
- Do not invent requirements.
- Do not perform opportunistic refactors.
- Do not modify secrets.
- Do not commit generated GitNexus indexes or local tool state.
- Run relevant checks or explain why they cannot be run.
- Update `docs/STATUS.md` when project state changes.
- Record durable lessons in `memory/LEARNINGS.md`.
- Record architecture decisions in `docs/DECISIONS.md`.

## Stop conditions

Stop and ask/report when:

- `memory/PROJECT_BRIEF.md` is empty or contradicts the task.
- GitNexus is required but not indexed/working.
- `.env` values resolve empty in `opencode debug config`.
- The requested work would spend OpenAI API credit without explicit approval.
- The task requires changes outside the stated scope.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **docker-image-watcher** (591 symbols, 856 relationships, 27 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "main"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/docker-image-watcher/context` | Codebase overview, check index freshness |
| `gitnexus://repo/docker-image-watcher/clusters` | All functional areas |
| `gitnexus://repo/docker-image-watcher/processes` | All execution flows |
| `gitnexus://repo/docker-image-watcher/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
