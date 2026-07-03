param(
    [Parameter(Mandatory=$true)]
    [string]$ProjectName,

    [Parameter(Mandatory=$false)]
    [string]$ForgejoRemote
)

$ErrorActionPreference = "Stop"

Get-ChildItem -Recurse -File |
    Where-Object { $_.FullName -notmatch '\\.git(\\|$)' -and $_.FullName -notmatch '\\.gitnexus(\\|$)' } |
    ForEach-Object {
        $content = Get-Content $_.FullName -Raw
        $content = $content -replace 'TODO_PROJECT', $ProjectName
        Set-Content -Path $_.FullName -Value $content -NoNewline
    }

if (-not (Test-Path ".git")) {
    git init
}

if ($ForgejoRemote) {
    git remote get-url origin *> $null
    if ($LASTEXITCODE -eq 0) {
        git remote set-url origin $ForgejoRemote
    } else {
        git remote add origin $ForgejoRemote
    }
}

Write-Host "Bootstrap complete for $ProjectName"
