# Caveman Handoff Format

Use this when work should be executed in Caveman Code after OpenCode planning.

## Preferred protocol

Use the `agent_bridge` MCP tools when they are available. Depending on the client UI, they may appear as bare `handoff_*` names or with an `agent_bridge_` prefix.

Before using broker tools in a session:

1. Verify which exact broker tool names are visible in that client.
2. Use those exact names for the rest of the session.
3. If broker tools are not visible, switch to fallback protocol below.

Preferred flow:

1. Plan the task in OpenCode first.
2. Use the exact visible create tool targeting `caveman`.
3. Tell the user to open Caveman and ask it to inspect pending handoffs.
4. After Caveman finishes, use the exact visible get tool with the returned `handoff_id`.
5. Plan the next step from the structured result.

## Fallback protocol

If `agent_bridge` tools are not available in the active OpenCode or Caveman session:

1. Write execution instructions to `tmp/handoff/current-task.md`.
2. Tell the user: `Ready. Run .\scripts\run-caveman.ps1`
3. Caveman reads `tmp/handoff/current-task.md` and executes.
4. Caveman writes `tmp/handoff/current-result.md`.
5. After Caveman finishes, the user tells OpenCode: `Read tmp/handoff/current-result.md`
6. OpenCode reads the result and plans the next step.

## Rules

- Prefer `agent_bridge` over file handoff when available.
- Use exact client-visible broker tool names, not guessed names.
- Treat `tmp/handoff/` as transient fallback scratch.
- Do not commit `tmp/handoff/current-task.md` or `tmp/handoff/current-result.md`.
- Keep tasks scoped and verifiable.
- Do not include template or agentic files in upstream PR branches.
- Each PR branch must be based on `upstream/main`.
- Each commit must pass the required verification for that task.

## Handoff payload guidance

When creating a handoff, include:

- target agent id (`caveman` or `opencode`);
- short task title;
- clear goal;
- files expected to change;
- ordered steps;
- required verification commands;
- constraints and branch expectations;
- metadata useful for traceability.
