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

Do not scatter application files across the repository root.

The repository root is reserved for:

- project instructions and documentation;
- tool configuration;
- environment examples;
- scripts;
- Git/Forgejo metadata;
- top-level build files only when the project genuinely requires them.

Default project code location:

```text
src/
```

Use `src/` for new application/source files unless `docs/ARCHITECTURE.md` defines another dedicated implementation folder such as:

```text
app/
packages/
services/
infra/
```

Temporary experiments, generated scratch work, and one-off agent notes must not be dropped in the root. Put durable tasks in `docs/TASKS.md` or `tasks/`; put temporary local scratch files under `tmp/` and do not commit them.

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
