param(
    [string]$EnvFile = ".env",
    [switch]$NoDebug
)

$ErrorActionPreference = "Stop"

$ProjectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $ProjectRoot

& "$PSScriptRoot\load-opencode-env.ps1" -EnvFile $EnvFile

if (-not $NoDebug) {
    Write-Host ""
    Write-Host "Checking OpenCode resolved config..."
    opencode debug config
}

Write-Host ""
Write-Host "Starting OpenCode..."
opencode
