# Agent Ledger Windows Installer (PowerShell)
# Repository: https://github.com/rahumanrahuu/agent-ledger
# Usage: irm https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.ps1 | iex

[CmdletBinding()]
param (
    [string]$Version = $env:VERSION
)

$ErrorActionPreference = "Stop"

$Repo = "rahumanrahuu/agent-ledger"

# Detect Windows architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
switch ($Arch) {
    "AMD64" { $TargetArch = "amd64" }
    "ARM64" { $TargetArch = "arm64" }
    default {
        Write-Error "Unsupported architecture: $Arch. Agent Ledger supports AMD64 and ARM64 on Windows."
        exit 1
    }
}

# Determine target version
if ($Version) {
    if (-not $Version.StartsWith("v")) {
        $Tag = "v$Version"
    } else {
        $Tag = $Version
    }
    Write-Host "Installing requested version: $Tag"
} else {
    Write-Host "Determining latest release for $Repo..."
    $Tag = $null
    try {
        $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{"Accept" = "application/vnd.github.v3+json"} -ErrorAction Stop
        if ($ReleaseInfo.tag_name) {
            $Tag = $ReleaseInfo.tag_name
        }
    } catch {
        # Fallback if API rate-limited
        $Tag = "v0.2.1"
    }

    if (-not $Tag) {
        $Tag = "v0.2.1"
    }
    Write-Host "Found latest version: $Tag"
}

$ArchiveName = "agent-ledger_${Tag}_windows_${TargetArch}.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$ArchiveName"

# Setup temporary directory
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

# Setup user-local installation directory
$InstallDir = if ($env:LOCALAPPDATA) {
    Join-Path $env:LOCALAPPDATA "Programs\agent-ledger"
} else {
    Join-Path $HOME ".local\bin"
}
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

try {
    $ZipPath = Join-Path $TempDir $ArchiveName
    Write-Host "Downloading $ArchiveName..."
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing

    Write-Host "Extracting binaries..."
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    $AgentLedgerExe = Join-Path $TempDir "agent-ledger.exe"
    $LedgerMcpExe = Join-Path $TempDir "ledger-mcp.exe"

    if (-not (Test-Path $AgentLedgerExe) -or -not (Test-Path $LedgerMcpExe)) {
        Write-Error "Release archive did not contain agent-ledger.exe and ledger-mcp.exe."
        exit 1
    }

    Write-Host "Installing executables to $InstallDir..."
    Copy-Item -Path $AgentLedgerExe -Destination (Join-Path $InstallDir "agent-ledger.exe") -Force
    Copy-Item -Path $LedgerMcpExe -Destination (Join-Path $InstallDir "ledger-mcp.exe") -Force

    # Verify installation
    $InstalledAgentLedger = Join-Path $InstallDir "agent-ledger.exe"
    $InstalledLedgerMcp = Join-Path $InstallDir "ledger-mcp.exe"

    if ((Test-Path $InstalledAgentLedger) -and (Test-Path $InstalledLedgerMcp)) {
        Write-Host ""
        Write-Host "Successfully installed Agent Ledger ($Tag)!" -ForegroundColor Green
        Write-Host "  - CLI: $InstalledAgentLedger"
        Write-Host "  - MCP: $InstalledLedgerMcp"
        Write-Host ""
    } else {
        Write-Error "Installation verification failed in $InstallDir."
        exit 1
    }

    # Check PATH
    $UserPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Host "Notice: $InstallDir is not currently in your User PATH." -ForegroundColor Yellow
        Write-Host "To add it automatically for your user account in PowerShell, run:"
        Write-Host ""
        Write-Host "  [System.Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$InstallDir', 'User')"
        Write-Host ""
        Write-Host "Or add '$InstallDir' to your User Environment Variables in Windows Settings."
    } else {
        Write-Host "Verification:"
        Write-Host "  Run 'agent-ledger --help' to get started."
        Write-Host "  Run 'ledger-mcp --help' to view MCP server details."
    }
} finally {
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
