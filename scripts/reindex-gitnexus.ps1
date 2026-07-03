param(
    [switch]$SkipGit
)

$ErrorActionPreference = "Stop"

if (-not $SkipGit -and -not (Test-Path ".git")) {
    throw "No .git directory found. Initialize Git first or rerun with -SkipGit for early setup."
}

if ($SkipGit) {
    gitnexus analyze --skip-git
} else {
    gitnexus analyze
}

gitnexus status
