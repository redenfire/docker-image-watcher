param(
    [switch]$SkipGitNexus,
    [switch]$SkipCaveman,
    [switch]$InstallSearchTools
)

$ErrorActionPreference = "Stop"

if ($InstallSearchTools) {
    winget install -e --id sharkdp.fd
    winget install -e --id BurntSushi.ripgrep.MSVC
}

if (-not $SkipGitNexus) {
    npm install -g gitnexus@latest
}

if (-not $SkipCaveman) {
    npm install -g @juliusbrussee/caveman-code
}

Write-Host "Agent workstation tools installation complete."
Write-Host "Remember: OpenCode must be installed separately according to docs/OPENCODE.md."
