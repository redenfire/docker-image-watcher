# Runbook

Operational commands for this project.

## Start OpenCode

Windows:

```powershell
.\scripts\start-opencode.ps1
```

Linux/macOS:

```bash
./scripts/start-opencode.sh
```

## Verify OpenCode environment

```powershell
opencode debug config
```

Provider keys and URLs must not be empty.

## Verify GitNexus

```powershell
gitnexus status
opencode mcp list
opencode mcp debug gitnexus
```

## Reindex GitNexus

```powershell
.\scripts\reindex-gitnexus.ps1
```

or:

```bash
./scripts/reindex-gitnexus.sh
```

For early setup without Git history:

```powershell
.\scripts\reindex-gitnexus.ps1 -SkipGit
```

or:

```bash
./scripts/reindex-gitnexus.sh --skip-git
```

## Check template/project scaffold

```powershell
.\scripts\check.ps1
```

or:

```bash
./scripts/check.sh
```
