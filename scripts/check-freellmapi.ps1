$ErrorActionPreference = "Stop"

if (Test-Path .env) {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*#') { return }
        if ($_ -match '^\s*$') { return }
        $parts = $_ -split '=', 2
        if ($parts.Length -eq 2) {
            [Environment]::SetEnvironmentVariable($parts[0].Trim(), $parts[1].Trim(), 'Process')
        }
    }
}

$BaseUrl = $env:FREELLMAPI_BASE_URL
if (-not $BaseUrl) { $BaseUrl = "http://127.0.0.1:3001/v1" }

$ApiKey = $env:FREELLMAPI_API_KEY
if (-not $ApiKey) {
    throw "FREELLMAPI_API_KEY is not set. Add it to .env or set it in the current PowerShell session."
}

$Headers = @{
    Authorization = "Bearer $ApiKey"
    "Content-Type" = "application/json"
}

Write-Host "Checking FreeLLMAPI models endpoint: $BaseUrl/models"
Invoke-RestMethod -Uri "$BaseUrl/models" -Headers $Headers -Method Get | Out-Null

Write-Host "Checking FreeLLMAPI chat endpoint with model=auto"
$Body = @{
    model = "auto"
    messages = @(@{ role = "user"; content = "Reply with OK only." })
    stream = $false
} | ConvertTo-Json -Depth 5
Invoke-RestMethod -Uri "$BaseUrl/chat/completions" -Headers $Headers -Method Post -Body $Body | Out-Null

Write-Host "FreeLLMAPI check passed."
