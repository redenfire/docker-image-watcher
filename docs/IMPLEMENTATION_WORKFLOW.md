# Implementation Workflow

This document defines the normal handoff from planning to implementation.

The template intentionally separates project analysis from implementation:

```text
OpenCode
= analyze project state, use GitNexus MCP, refine task scope, produce an approved implementation prompt

Caveman Code
= verify GitNexus access, execute the approved prompt as the focused coding endpoint, and use Codex/ChatGPT-style authentication when configured
```

## Repository layout rule

This repository intentionally preserves the upstream application layout at the repository root:

```text
main.go
docker.go
registry.go
web/
Dockerfile
docker-compose.yml
```

Do not invent a `src/` migration or scatter additional root-level files beyond the established upstream app layout plus project instructions/docs, tool configuration, environment examples, scripts, Git/Forgejo metadata, and genuinely required top-level build files.

Temporary experiments, generated scratch work, and one-off agent notes must not be dropped in the root. Put durable tasks in `docs/TASKS.md` or `tasks/`; put temporary local scratch files under `tmp/` and do not commit them.

## Branch policy for upstream contribution work

Use two branch types on purpose:

- `main` = private Forgejo integration branch for this repository's accepted local tooling, docs, and upstream-synced application state;
- `pr/*` = disposable upstream contribution branches cut from fresh `upstream/main`.

Rules:

1. Do not open upstream PRs from private `main`.
2. Start each upstreamable fix branch from latest `upstream/main`, then apply only the minimal upstreamable patch.
3. Do not include local-only files in any `pr/*` branch, including `.env`, `opencode.json`, runtime handoff archives, scratch notes, or generated local state under `tmp/`.
4. If upstream merges an equivalent fix or moves significantly, delete stale `pr/*` branches and recreate them from latest `upstream/main` instead of stacking new rebases onto old review branches.

## Required OpenCode planning pass

Before Caveman Code implementation, run OpenCode in planning mode.

OpenCode must:

1. read the required project documents;
2. use GitNexus MCP for codebase navigation;
3. detect project-readiness issues;
4. refine task scope;
5. identify the correct implementation folder;
6. produce a Caveman Code-ready implementation prompt;
7. avoid editing files.

## OpenCode planning prompt

Use this prompt at the start of real work:

```text
Read the project instructions and required project documents:

- AGENTS.md
- docs/HOW_TO_USE.md
- docs/TOOLING_MODEL.md
- docs/IMPLEMENTATION_WORKFLOW.md
- docs/STATUS.md
- docs/TASKS.md
- docs/ROADMAP.md
- docs/ARCHITECTURE.md
- memory/PROJECT_BRIEF.md
- memory/CONSTRAINTS.md

Use GitNexus MCP for repository navigation and codebase understanding.

Enter plan mode.

Analyze the current project state and prepare a precise execution plan for TASK-001.

Do not edit files yet.

Your output must include:

1. task interpretation;
2. project-readiness issues, if any;
3. expected implementation folder, usually src/ unless ARCHITECTURE.md says otherwise;
4. files likely involved;
5. implementation steps;
6. checks/tests to run;
7. risks and assumptions;
8. exact instructions suitable to paste into Caveman Code for implementation.

Keep the plan scoped to TASK-001.
Do not introduce unrelated refactors.
Do not start implementation.
```

## Caveman Code execution prompt

After reviewing and approving the OpenCode plan, paste the final implementation instructions into Caveman Code with this wrapper:

```text
Read AGENTS.md, docs/TOOLING_MODEL.md, docs/IMPLEMENTATION_WORKFLOW.md, docs/CAVEMAN_GITNEXUS.md, and the approved OpenCode execution plan below.

Before implementation, verify that GitNexus is available in this Caveman Code session. Prefer GitNexus MCP tools. If GitNexus MCP tools are not visible, check whether `gitnexus status` works from the project root and explicitly report that you are using CLI fallback. If neither works, stop and report the blocker.

Implement only the approved TASK-001 scope.

Do not perform unrelated refactors.
Do not modify secrets.
Do not modify generated indexes.
Do not scatter new files in the repository root.
Place new implementation files under the dedicated implementation folder identified by the plan, usually src/ unless docs/ARCHITECTURE.md says otherwise.
Use the smallest correct change.

After implementation:

1. state whether GitNexus MCP or GitNexus CLI fallback was used;
2. list changed files;
3. run the requested checks;
4. report check results;
5. explain any deviations from the plan;
6. update docs/STATUS.md only if the project state changed.

Approved OpenCode plan:

[PASTE OPENCODE PLAN HERE]
```

## Review after implementation

After Caveman Code finishes:

1. inspect changed files;
2. verify files are in the expected implementation folder;
3. run checks manually if needed;
4. update `docs/STATUS.md` and `memory/LEARNINGS.md` if durable project state changed;
5. commit only a coherent task-sized change.
