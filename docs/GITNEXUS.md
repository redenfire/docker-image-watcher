# GitNexus

GitNexus is required. It is the repository memory/index layer used by OpenCode through MCP and by Caveman Code when the Cave-side config/session exposes it.

Do not treat GitNexus as optional in this template.

## Install

Default install:

```powershell
npm install -g gitnexus@latest
```

## Initialize the repo

Preferred normal path:

```powershell
git init -b main
git add .
git commit -m "Initial project import"
gitnexus analyze
gitnexus status
```

Early setup path when Git history is not ready:

```powershell
gitnexus analyze --skip-git
gitnexus status
```

Use `--skip-git` as an early bootstrap escape hatch, not as a permanent replacement for a real Git repository.

## Windows parser workaround tested in this project

Some GitNexus Tree-sitter grammar packages are fragile under Windows/Node combinations.

Observed errors included:

- missing `tree-sitter-kotlin`;
- no native build for `tree-sitter-dart`;
- missing `tree-sitter-swift`.

Tested practical path:

```powershell
npm uninstall -g gitnexus
npm cache verify
$env:GITNEXUS_SKIP_OPTIONAL_GRAMMARS = "1"
npm install -g gitnexus@latest
npm install -g tree-sitter-dart tree-sitter-swift
```

Then:

```powershell
gitnexus analyze --skip-git
gitnexus status
```

For normal projects after Git is initialized:

```powershell
gitnexus analyze
gitnexus status
```

Do not install random parser packages one by one unless the error explicitly names the missing parser. The two packages above were required in testing because GitNexus still imported them at runtime.

## Endpoint configuration

OpenCode and Caveman Code use different configuration files.

OpenCode:

```text
opencode.json -> mcp.gitnexus
```

Caveman Code / Cave:

```text
.cave/settings.json -> mcp.gitnexus
```

Do not assume OpenCode MCP configuration is inherited by Caveman Code. See `docs/CAVEMAN_GITNEXUS.md`.

## MCP verification

GitNexus must expose MCP to OpenCode, and Caveman Code must verify GitNexus MCP tools or an explicit CLI fallback before repository-wide implementation work.

```powershell
opencode mcp list
opencode mcp debug gitnexus
```

If MCP fails:

1. Check `gitnexus status`.
2. Check `gitnexus mcp` starts manually.
3. Check `opencode debug config`.
4. Confirm OpenCode was launched from the project root through the env loader.
5. Increase MCP timeout if startup is slow.

## Agent rule

Agents should use GitNexus before broad file exploration. If GitNexus is unavailable in the active endpoint, report the blocker instead of silently falling back to guesswork. Caveman Code must not claim OpenCode MCP access as its own; it must verify Cave-side tools or report CLI fallback.
