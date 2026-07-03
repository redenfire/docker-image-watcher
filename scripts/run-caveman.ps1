<#
.SYNOPSIS
    Launch Caveman Code with current handoff context.
.DESCRIPTION
    Reads tmp/handoff/current-task.md, prints the exact Caveman instruction,
    optionally copies it to the clipboard, launches Caveman Code, then reminds
    the user to have OpenCode read tmp/handoff/current-result.md after Caveman
    finishes.
#>

$projectRoot = Split-Path -Parent $PSScriptRoot
$handoffDir = Join-Path $projectRoot "tmp\handoff"
$taskFile = Join-Path $handoffDir "current-task.md"
$resultFile = Join-Path $handoffDir "current-result.md"

if (-not (Test-Path $taskFile)) {
    Write-Host "==============================================================" -ForegroundColor Red
    Write-Host "ERROR: tmp/handoff/current-task.md not found" -ForegroundColor Red
    Write-Host "Have OpenCode write the handoff task first." -ForegroundColor Red
    Write-Host "==============================================================" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $handoffDir)) {
    New-Item -ItemType Directory -Force -Path $handoffDir | Out-Null
}

$instruction = "Read tmp/handoff/current-task.md and execute. Write results to tmp/handoff/current-result.md."
$clipboardReady = $false

try {
    if (Get-Command Set-Clipboard -ErrorAction Stop) {
        Set-Clipboard -Value $instruction
        $clipboardReady = $true
    }
} catch {
    $clipboardReady = $false
}

Write-Host "==============================================================" -ForegroundColor Cyan
Write-Host "Caveman Code Handoff" -ForegroundColor Cyan
Write-Host "--------------------------------------------------------------" -ForegroundColor Cyan
Write-Host "Task file:   tmp/handoff/current-task.md" -ForegroundColor Cyan
Write-Host "Result file: tmp/handoff/current-result.md" -ForegroundColor Cyan
Write-Host "--------------------------------------------------------------" -ForegroundColor Cyan
Write-Host "Tell Caveman exactly this:" -ForegroundColor Yellow
Write-Host "" -ForegroundColor Yellow
Write-Host $instruction -ForegroundColor Yellow
if ($clipboardReady) {
    Write-Host "" -ForegroundColor Yellow
    Write-Host "Instruction copied to clipboard." -ForegroundColor Yellow
}
Write-Host "==============================================================" -ForegroundColor Cyan
Write-Host ""

caveman

Write-Host ""
Write-Host "==============================================================" -ForegroundColor Green
Write-Host "Caveman session ended." -ForegroundColor Green
Write-Host "If results were written, tell OpenCode:" -ForegroundColor Green
Write-Host "  Read tmp/handoff/current-result.md" -ForegroundColor Green
Write-Host "==============================================================" -ForegroundColor Green
