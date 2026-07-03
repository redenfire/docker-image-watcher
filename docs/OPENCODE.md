# OpenCode

OpenCode is the primary configurable local agent interface for this template. It is separate from VS Code, Caveman Code, and Codex in VS Code.

Install OpenCode once per workstation/user. The project repository provides configuration and prompts; it does not vendor OpenCode.


## Boundary with other tools

OpenCode controls only the OpenCode agent session.

- VS Code is the editor/workbench.
- OpenCode uses the model routes in `opencode.json`: FreeLLMAPI, DeepSeek, and optional OpenAI API.
- GitNexus is exposed to OpenCode through MCP using `opencode.json`.
- Caveman Code is a separate terminal agent and does not use `opencode.json`; it uses `.cave/settings.json` for project-side GitNexus MCP configuration when supported.
- Codex in VS Code is optional and does not use OpenCode routing.

Do not expect OpenCode `/models` to change Caveman Code, and do not expect Caveman Code login to change OpenCode authentication.

## Required launch rule

Do not start OpenCode directly after editing `.env`.

`.env` is not loaded automatically by Windows, Linux, macOS, PowerShell, Bash, or OpenCode.

Use:

```powershell
.\scripts\start-opencode.ps1
```

or:

```bash
./scripts/start-opencode.sh
```

Then inspect:

```powershell
opencode debug config
```

Provider `apiKey` and `baseURL` values must not be empty.

## Providers vs models vs agents

| Concept | Meaning | How to change |
|---|---|---|
| Provider | API/backend definition | `provider` block in `opencode.json` |
| Main model | Default model for normal prompts | `model` in `opencode.json` or `/models` |
| Small model | Lightweight/background model | `small_model` in `opencode.json` |
| Agent | Behavior/persona/tool permission profile | TAB cycles primary agents |

TAB cycles agents, not providers. Use `/models` to switch the active model route inside OpenCode.

## Default routing

```json
"model": "deepseek/deepseek-chat",
"small_model": "freellmapi/auto"
```


FreeLLMAPI model id note:

- use `freellmapi/auto`;
- do not use `freellmapi/free-auto`;
- confirm the live FreeLLMAPI catalog with `/v1/models` if model selection fails.

Use:

- DeepSeek for normal planning and coding.
- FreeLLMAPI for low-risk/cheap route and small-model work.
- OpenAI API only when explicitly selected and intentionally funded.

## OpenAI API selection/deselection

OpenAI provider presence does not mean OpenAI is selected.

OpenAI API is deselected when:

- `model` does not point to `openai/...`;
- `small_model` does not point to `openai/...`;
- `/models` has not selected OpenAI;
- `OPENAI_API_KEY` is not loaded.

OpenAI API is selected only when:

- `OPENAI_API_KEY` is loaded; and
- `model`, `small_model`, or `/models` points to `openai/...`; and
- the user explicitly approves OpenAI API spend.

Business ChatGPT/Codex usage is separate from OpenAI API-key billing.

## MCP is required

OpenCode must always be configured with MCP.

Required MCP:

```json
"mcp": {
  "gitnexus": {
    "type": "local",
    "command": ["gitnexus", "mcp"],
    "enabled": true,
    "timeout": 15000
  }
}
```

GitNexus tools are disabled globally and enabled per agent:

```json
"tools": {
  "gitnexus_*": false
}
```

This prevents uncontrolled tool use while allowing plan/build/review agents to use GitNexus explicitly.

## MCP verification checklist

Before serious OpenCode work:

```powershell
gitnexus status
opencode mcp list
opencode mcp debug gitnexus
```

Inside OpenCode:

```text
Use GitNexus tools to summarize this repository.
```

MCP is not working merely because it appears in `opencode.json`; it is working only when `opencode mcp debug gitnexus` succeeds and the agent can call `gitnexus_*` tools. Caveman Code does not read `opencode.json`; its counterpart is `.cave/settings.json`.

## Debug-first flow

1. Run `scripts/start-opencode.*`.
2. Check `opencode debug config`.
3. Confirm env values are non-empty.
4. Confirm `opencode mcp debug gitnexus`.
5. Start with plan agent.
6. Use `/models` only when intentionally changing model route.


## Planning before implementation

For serious project work, OpenCode should plan before any coding endpoint edits files.

OpenCode should:

1. read `AGENTS.md` and required project docs;
2. use GitNexus MCP for repository navigation;
3. inspect `docs/ARCHITECTURE.md` to identify the implementation folder;
4. produce a scoped plan for the active task;
5. include exact instructions suitable for Caveman Code.

The default implementation folder is `src/` unless the architecture defines another dedicated location.

See `docs/IMPLEMENTATION_WORKFLOW.md`.
