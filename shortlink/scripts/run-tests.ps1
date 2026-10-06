```powershell
$ErrorActionPreference = "Stop"

Set-Location (Split-Path $PSScriptRoot -Parent)

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$reportDir = Join-Path (Get-Location) "reports"
$reportFile = Join-Path $reportDir "test-$timestamp.log"

New-Item -ItemType Directory -Force -Path $reportDir | Out-Null

function Invoke-TestStep {
    param(
        [string]$Title,
        [string]$Command,
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "========================================"
    Write-Host $Title
    Write-Host "========================================"

    Add-Content -Path $reportFile -Value "`r`n===== $Title ====="

    & $Action 2>&1 |
        Tee-Object -FilePath $reportFile -Append

    if ($LASTEXITCODE -ne 0) {
        throw "$Title failed. Command: $Command"
    }
}

Invoke-TestStep `
    -Title "Go Unit Tests" `
    -Command "go test ./..." `
    -Action { go test ./... }

Invoke-TestStep `
    -Title "Go Race Detector" `
    -Command "go test -race ./..." `
    -Action { go test -race ./... }

Invoke-TestStep `
    -Title "Singleflight Benchmark" `
    -Command "go test ./internal/service -run '^$' -bench 'BenchmarkSingleflight' -benchmem -count=3" `
    -Action {
        go test ./internal/service `
            -run '^$' `
            -bench 'BenchmarkSingleflight' `
            -benchmem `
            -count=3
    }

Write-Host ""
Write-Host "All test steps passed."
Write-Host "Report: $reportFile"
```