# Agent Bridge MCP

`agent_bridge` is a local MCP server that lets OpenCode and Caveman Code exchange structured work items through a shared task bus instead of manual copy/paste.

## Scope

Phase 1 provides shared queue tools only. It does not auto-launch Caveman Code or OpenCode.

That means this is already true tool interoperability, but not full orchestration yet.

## Architecture

```text
OpenCode                         Caveman Code
    |                                 |
    |  MCP tools                      |  MCP tools
    +--------> agent_bridge <---------+
                   |
             tmp/agent-bridge/
               queue/
               archive/
               locks/
```

Both agents register the same local MCP server under the `agent_bridge` key:

- OpenCode via local-only `opencode.json`
- Caveman/Cave via `.cave/settings.json`

OpenCode local config snippet:

```json
"agent_bridge": {
  "type": "local",
  "command": ["node", "tools/agent-bridge/server.mjs"],
  "enabled": true,
  "timeout": 15000
}
```

Cave/Caveman config snippet:

```json
"agent_bridge": {
  "type": "local",
  "command": ["node", "tools/agent-bridge/server.mjs"],
  "enabled": true,
  "timeout": 15000
}
```

## Runtime storage

Runtime state lives under `tmp/agent-bridge/`:

- `queue/` — active handoffs
- `archive/` — completed or failed handoffs
- `locks/` — reserved for future coordination work

Handoffs are stored as JSON, not Markdown, so both agents can read and update state deterministically.

## State model

```text
queued -> claimed -> completed
                  -> failed
```

`in_progress` and `cancelled` are reserved for later phases and are recognized as future-compatible states.

## Tools

Use the `agent_bridge` MCP tools. Depending on client UI, they may appear as bare `handoff_*` names or with an `agent_bridge_` prefix.

| Canonical action | Possible client-visible names | Called by | Purpose |
|---|---|---|---|
| create | `handoff_create` or `agent_bridge_handoff_create` | OpenCode | Queue work for another agent |
| list | `handoff_list` or `agent_bridge_handoff_list` | Caveman or OpenCode | List active handoffs for an agent |
| claim | `handoff_claim` or `agent_bridge_handoff_claim` | Caveman or OpenCode | Claim one queued handoff |
| get | `handoff_get` or `agent_bridge_handoff_get` | Both | Read full task or archived result |
| complete | `handoff_complete` or `agent_bridge_handoff_complete` | Claiming agent | Archive successful result |
| fail | `handoff_fail` or `agent_bridge_handoff_fail` | Claiming agent | Archive failure result |

## Finding exact tool names in each client

Use the exact tool names exposed by the active client session, not assumptions from docs.

### OpenCode

Verify MCP visibility first:

```powershell
opencode mcp list
```

If broker is loaded, use whichever names OpenCode exposes for the six handoff actions above.

### Caveman/Cave

Start a fresh session after config changes. Then inspect available MCP tools in-session and use the exact visible names for list, claim, get, complete, and fail.

If the session shows no `agent_bridge` or `handoff_*` tools, broker is not active in that endpoint yet. Fall back to file handoff or fix MCP loading first.

## Typical workflow

1. OpenCode plans work.
2. OpenCode uses the exact visible create tool for the current session and targets `caveman`.
3. User opens Caveman and asks it to inspect pending handoffs.
4. Caveman uses the exact visible list tool, then the exact visible claim tool.
5. Caveman implements the task and runs verification.
6. Caveman uses the exact visible complete or fail tool.
7. User returns to OpenCode.
8. OpenCode uses the exact visible get tool to read result and plan next step.

## Setup

Install broker dependency from repo root:

```powershell
cd tools/agent-bridge
npm install
cd ../..
```

Optional quick syntax check:

```powershell
node --check tools/agent-bridge/server.mjs
```

Manual smoke test:

```powershell
node tools/agent-bridge/server.mjs
```

The server will wait on stdio for an MCP client. Stop it with `Ctrl+C` if launched manually.

## Notes

- Keep `tmp/agent-bridge/` out of commits except for `.gitignore`.
- This broker does not replace GitNexus. GitNexus remains required for repository understanding and impact analysis.
- Broker writes use atomic temp-file rename and per-handoff lock directories with stale-lock expiry to reduce concurrent mutation races.
- Supported agent ids are currently `caveman` and `opencode`.
- MCP hosts must launch repo-local commands from a project-aware session so `node tools/agent-bridge/server.mjs` resolves from repository context.
- The existing `scripts/run-caveman.ps1` helper remains a fallback path when MCP broker tools are unavailable in a session.
