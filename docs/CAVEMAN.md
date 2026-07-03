# Caveman

This template distinguishes two related tools. See `docs/TOOLING_MODEL.md` for how Caveman fits beside VS Code and OpenCode.

## Caveman Code

Caveman Code is a standalone terminal coding agent. It is not controlled by `opencode.json`, and OpenCode model routing does not affect it. Cave/Caveman must use its own project-side configuration, including `.cave/settings.json` for GitNexus MCP.

Install once per workstation/user:

```powershell
npm install -g @juliusbrussee/caveman-code
caveman --version
caveman-code --version
```

Launch from the project root:

```powershell
caveman
```

Login:

```text
/login
```

Follow the interactive prompt. Do not document `caveman login --provider openai` as the ChatGPT login path; `openai` is API-key provider terminology, not the tested ChatGPT subscription login flow.

Use Caveman Code when:

- you want a separate low-token terminal coding endpoint;
- you want to use ChatGPT/Codex auth through Caveman Code;
- you do not need the VS Code Codex UI;
- you want a separate endpoint from OpenCode while still working in the same repository.

## Windows Node note

If Caveman Code install fails with `better-sqlite3` / `node-gyp` on Node Current/Node 26:

1. Uninstall Node Current/Node 26.
2. Install Node.js LTS.
3. Open a new PowerShell.
4. Reinstall Caveman Code globally.

## GitNexus access

Caveman Code should not be blind to the repository. This template includes a Cave-side GitNexus MCP config in:

```text
.cave/settings.json
```

That file mirrors the OpenCode GitNexus MCP command:

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

Important: `opencode.json` is not read by Caveman Code. If GitNexus works in OpenCode but not in Cave, the Cave-side config/session layer is the problem, not OpenCode.

Before implementation, Caveman Code should verify GitNexus access. If no GitNexus MCP tools are visible, it may use `gitnexus status` / GitNexus CLI only when the user accepts that fallback. Otherwise it must stop and report the missing memory/index layer.

See `docs/CAVEMAN_GITNEXUS.md`.

## Caveman Skill

Caveman Skill is separate from Caveman Code.

It is a brevity/output-control integration installed into specific agents/editors.

Install it only if you want Caveman behavior inside another agent such as Codex, Claude Code, Gemini, or OpenCode.

Example:

```powershell
irm https://raw.githubusercontent.com/JuliusBrussee/caveman/main/install.ps1 | iex
```

or targeted installer if supported:

```powershell
npx -y github:JuliusBrussee/caveman -- --only codex
```

Restart the relevant editor/agent after installing.

## Difference

| Tool | Role | Installed how |
|---|---|---|
| Caveman Code | Standalone terminal coding agent | system/user-wide npm package |
| Caveman Skill | Brevity skill/plugin for other agents | per-agent/editor integration |

Installing Caveman Code does not automatically add `/caveman` or `$caveman` inside Codex. Install Caveman Skill only if you need that agent/editor integration.


## Executing an OpenCode-approved plan

The normal project workflow is not “ask Caveman Code to figure everything out from scratch.”

Use OpenCode first to read the project docs, use GitNexus MCP, and produce a scoped implementation plan. Then paste the approved implementation instructions into Caveman Code. Caveman Code should also use GitNexus through `.cave/settings.json` when Cave exposes those MCP tools.

Caveman Code execution rules:

- verify GitNexus MCP tools or accepted GitNexus CLI fallback before broad repository work;
- implement only the approved task scope;
- do not perform unrelated refactors;
- do not modify secrets;
- do not modify generated GitNexus or local tool state;
- do not scatter source files in the repository root;
- place new implementation files under `src/` unless `docs/ARCHITECTURE.md` defines another dedicated folder;
- run requested checks and report changed files.

See `docs/IMPLEMENTATION_WORKFLOW.md` for the copy/paste wrapper prompt.
