$ErrorActionPreference = "Stop"

$required = @(
    "README.md",
    "AGENTS.md",
    "docs/TOOLING_MODEL.md",
    "docs/HOW_TO_USE.md",
    "docs/IMPLEMENTATION_WORKFLOW.md",
    "docs/OPENCODE.md",
    "docs/GITNEXUS.md",
    "docs/CAVEMAN.md",
    "docs/CAVEMAN_GITNEXUS.md",
    "docs/CODEX.md",
    "docs/FREELLMAPI.md",
    "docs/STATUS.md",
    "docs/TASKS.md",
    "memory/PROJECT_BRIEF.md",
    "memory/CONSTRAINTS.md",
    "scripts/start-opencode.sh",
    "scripts/start-opencode.ps1",
    ".env.example",
    ".cave/settings.json",
    "tmp/agent-bridge/.gitignore",
    "tmp/handoff/.gitignore",
    "tools/agent-bridge/package.json"
)

foreach ($file in $required) {
    if (-not (Test-Path $file)) {
        throw "Missing required file: $file"
    }
}


if (Test-Path "opencode.json") {
    Get-Content opencode.json | ConvertFrom-Json | Out-Null
}

Write-Host "Scaffold check passed."
