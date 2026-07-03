# Tooling Model

This project intentionally uses more than one editor/agent endpoint. The goal is not to make every tool do the same job. Each tool has a separate role.

## Short version

```text
VS Code
= human editor and project workbench

OpenCode
= configurable local agent used with FreeLLMAPI, DeepSeek, optional OpenAI API, and GitNexus MCP

Caveman Code
= separate terminal coding agent used with Caveman's low-token workflow and ChatGPT/Codex-style authentication when configured

GitNexus
= required repository memory/index layer, exposed through MCP to OpenCode and, when Cave loads `.cave/settings.json`, to Caveman Code

FreeLLMAPI
= shared free/low-cost model gateway, normally used as OpenCode small_model or selected through /models

DeepSeek
= default cheap frontier model route for normal OpenCode work

OpenAI API
= backup/emergency OpenCode route only, explicitly selected and intentionally funded

Codex in VS Code
= optional IDE-integrated Codex endpoint, not required when Caveman Code covers the Codex/ChatGPT workflow

Caveman Skill
= optional brevity/plugin integration for specific agents/editors; separate from Caveman Code
```

## Canonical endpoint map

| Endpoint | Primary role | Uses | Required? |
|---|---|---|---|
| VS Code | Main human editor/workbench | Files, terminal, Git UI, optional extensions | Yes |
| OpenCode | Main configurable local agent | DeepSeek, FreeLLMAPI, optional OpenAI API, GitNexus MCP | Yes |
| Caveman Code | Separate terminal coding agent | Caveman low-token workflow, ChatGPT/Codex auth if configured, GitNexus via `.cave/settings.json` when supported | Yes/recommended |
| Codex in VS Code | Optional IDE-integrated Codex | ChatGPT/Codex account inside VS Code | Optional |
| Caveman Skill | Optional brevity integration | Installs Caveman behavior into selected agents/editors | Optional |

## OpenCode role

OpenCode is the normal project agent.

Use it for:

- project planning;
- routine implementation;
- documentation updates;
- repository-aware work through GitNexus MCP;
- controlled model routing through `opencode.json`.

Default OpenCode model routing:

```json
"model": "deepseek/deepseek-chat",
"small_model": "freellmapi/auto"
```

Meaning:

- DeepSeek is the normal cheap frontier route.
- FreeLLMAPI is the low-risk/small route.
- OpenAI API is not used unless explicitly selected.

In OpenCode:

- `TAB` cycles agents, not providers.
- `/models` changes the active model route.
- `opencode.json` sets persistent defaults.
- GitNexus MCP must be verified before serious OpenCode work.

## GitNexus access by endpoint

GitNexus is not an OpenCode-only tool. It is the required repository memory/index layer.

OpenCode gets GitNexus through `opencode.json`:

```text
opencode.json -> mcp.gitnexus -> gitnexus mcp
```

Caveman Code/Cave must be configured separately because it does not read `opencode.json`:

```text
.cave/settings.json -> mcp.gitnexus -> gitnexus mcp
```

Therefore, a working OpenCode MCP setup does not prove that Caveman Code has GitNexus access. Caveman Code must verify GitNexus tools or CLI access before broad repository work. See `docs/CAVEMAN_GITNEXUS.md`.

## Caveman Code role

Caveman Code is not an OpenCode plugin. It is a separate terminal coding endpoint.

Use it for:

- low-token coding sessions;
- focused terminal work;
- ChatGPT/Codex-authenticated coding when configured in Caveman Code;
- implementation of an OpenCode-approved plan;
- repository-aware implementation when Cave exposes GitNexus MCP from `.cave/settings.json`;
- cases where you do not need the VS Code Codex UI.

Launch from the project root:

```powershell
caveman
```

Login from inside the Caveman TUI:

```text
/login
```

## Codex role

Codex/ChatGPT access can happen through two paths:

1. Caveman Code with ChatGPT/Codex authentication.
2. Codex in VS Code, if you want IDE-integrated Codex.

Codex in VS Code is optional in this template. If Caveman Code is already authenticated and sufficient, use Caveman Code directly.

Do not confuse Codex/ChatGPT subscription usage with OpenAI API billing. OpenAI API usage inside OpenCode is a separate paid route controlled by `OPENAI_API_KEY` and `model` / `small_model` selection.

## Caveman Skill role

Caveman Skill is separate from Caveman Code.

Use Caveman Skill only when you specifically want Caveman brevity behavior inside another agent/editor, such as Codex, Claude Code, Gemini, or OpenCode.

Installing Caveman Code globally does not install `/caveman`, `$caveman`, or a Caveman function inside Codex.

## Anti-confusion rules

- VS Code is the editor, not the model router.
- OpenCode is the main configurable project agent.
- Caveman Code is a separate terminal coding agent.
- Codex in VS Code is optional, not the default Codex path.
- FreeLLMAPI and DeepSeek are OpenCode model routes.
- GitNexus is the repository memory layer, not a model provider.
- `opencode.json` configures OpenCode; `.cave/settings.json` configures the Cave/Caveman project side.
- Caveman Skill is an optional plugin/skill, not the same thing as Caveman Code.
- `.env` is not loaded automatically; use the provided OpenCode start scripts.


## OpenCode to Caveman Code handoff

The normal implementation pipeline is:

```text
OpenCode analyzes and plans -> user approves -> Caveman Code implements
```

OpenCode is used first because it has the project configuration, verified GitNexus MCP memory/index layer, and explicit model routing. Its job is to perfect the task plan and produce implementation instructions.

Caveman Code is then used as the focused coding endpoint, especially when ChatGPT/Codex-style authentication is configured. Its job is to implement the approved plan, run checks, and report changed files. Caveman Code should also use GitNexus when its Cave-side MCP configuration is loaded; it must not assume OpenCode MCP settings apply to Cave sessions.

Do not confuse this with editor choice: VS Code remains the human workbench, while Caveman Code runs from the terminal.
