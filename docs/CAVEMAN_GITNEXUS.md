# Caveman Code and GitNexus

GitNexus is the required repository memory/index layer for this repository.

OpenCode and Caveman Code are separate endpoints. Caveman Code does not read `opencode.json`, so the OpenCode MCP block is not enough for Cave/Caveman sessions.

## Required project-side Cave config

This repository includes:

```text
.cave/settings.json
```

as the canonical Cave-side config, plus a compatibility mirror in:

```text
.cave/mcp.json
```

Both should stay aligned for GitNexus and any shared project-local MCP servers such as `agent_bridge`.

`settings.json` carries the primary GitNexus MCP definition:

```json
{
  "mcp": {
    "gitnexus": {
      "type": "local",
      "command": ["gitnexus", "mcp"],
      "enabled": true,
      "timeout": 15000
    }
  }
}
```

This is the Cave-side counterpart to `opencode.json`.

## Verification rule

Do not assume GitNexus is available inside Caveman Code just because it works in OpenCode.

Before asking Caveman Code to implement a task, verify one of these is true:

1. the Caveman/Cave session exposes GitNexus MCP tools; or
2. the session can explicitly use the `gitnexus` CLI from the project root and the user accepts CLI fallback.

If neither is true, Caveman Code must stop and report that GitNexus is not available in that session.

## Manual checks

From the project root:

```powershell
gitnexus status
```

OpenCode-specific MCP check:

```powershell
opencode mcp debug gitnexus
```

Caveman Code check, inside Cave/Caveman:

```text
Check whether GitNexus MCP tools are available in this session. If they are not visible, check whether you can run `gitnexus status` from the project root. Do not proceed with broad repository work until one of those checks succeeds.
```

## Important distinction

`opencode.json` controls OpenCode.

`.cave/settings.json` is the primary project-side Cave/Caveman configuration location used by this repo, while `.cave/mcp.json` is kept as a compatibility mirror for clients that still read that file.

Changing one does not automatically change the other, so keep both aligned when adjusting repo-side MCP entries.

## Failure mode

If Caveman Code says only built-in functions are visible and no `gitnexus_*` tools are exposed, the likely causes are:

- Cave/Caveman did not load `.cave/settings.json`;
- the Cave API/session wrapper did not bind MCP tools into the session;
- GitNexus is not installed on PATH;
- GitNexus has not been indexed;
- the session is running outside the project root.

In that state, use OpenCode for GitNexus-heavy analysis, or fix the Cave-side MCP/plugin/session configuration before asking Caveman Code to perform repository-wide work.
