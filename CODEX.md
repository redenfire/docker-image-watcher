# Codex Instructions

Read `docs/TOOLING_MODEL.md` before assuming which editor, agent, provider, or model route is responsible for a task.


Use `AGENTS.md` as the controlling project policy.

Codex/ChatGPT usage is separate from OpenAI API-key billing.

Preferred access paths:

- Caveman Code for terminal low-token ChatGPT/Codex workflow.
- VS Code Codex only if IDE-integrated workflow is desired.

Do not assume GitNexus is optional. Before broad repo work, check that GitNexus indexing/MCP is available or report the blocker.


## Implementation layout

Do not scatter new source files in the repository root. Use `src/` unless `docs/ARCHITECTURE.md` defines another dedicated implementation folder. For the preferred flow, OpenCode produces the plan and Caveman Code executes the approved implementation instructions.
