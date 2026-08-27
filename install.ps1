# Agent Ledger Windows Installer (PowerShell)
# Repository: https://github.com/rahumanrahuu/agent-ledger
# Usage: irm https://raw.githubusercontent.com/rahumanrahuu/agent-ledger/main/install.ps1 | iex

[CmdletBinding()]
param (
    [string]$Version = $env:VERSION
)

$ErrorActionPreference = "Stop"

$Repo = "rahumanrahuu/agent-ledger"
# Known-good tags with published assets, used only if the GitHub API is unreachable.
$FallbackTags = @("v0.2.2", "v0.2.0")

# Ensure TLS 1.2+ on older PowerShell hosts
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
} catch { }

function Get-LatestTag {
    $Headers = @{ "Accept" = "application/vnd.github.v3+json"; "User-Agent" = "agent-ledger-installer" }
    foreach ($Endpoint in @(
        "https://api.github.com/repos/$Repo/releases/latest",
        "https://api.github.com/repos/$Repo/releases?per_page=10"
    )) {
        try {
            $Response = Invoke-RestMethod -Uri $Endpoint -Headers $Headers -TimeoutSec 15 -ErrorAction Stop
            if ($Response -is [array]) { $Candidates = $Response | ForEach-Object { $_.tag_name } }
            else { $Candidates = @($Response.tag_name) }
            foreach ($Candidate in $Candidates) {
                if ($Candidate) {
                    # Verify an asset actually exists before trusting this tag
                    $ProbeName = "agent-ledger_${Candidate}_windows_amd64.zip"
                    try {
                        Invoke-WebRequest -Uri "https://github.com/$Repo/releases/download/$Candidate/$ProbeName" -Method Head -Headers @{ "User-Agent" = "agent-ledger-installer" } -TimeoutSec 15 -UseBasicParsing | Out-Null
                        return $Candidate
                    } catch { continue }
                }
            }
        } catch { continue }
    }
    return $null
}

function Test-AssetExists {
    param([string]$Tag, [string]$Archive)
    try {
        Invoke-WebRequest -Uri "https://github.com/$Repo/releases/download/$Tag/$Archive" -Method Head -Headers @{ "User-Agent" = "agent-ledger-installer" } -TimeoutSec 20 -UseBasicParsing | Out-Null
        return $true
    } catch {
        return $false
    }
}

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

    # Auto-fallback to the latest available release if the requested one has no assets
    $RequestedArchive = "agent-ledger_${Tag}_windows_${TargetArch}.zip"
    if (-not (Test-AssetExists -Tag $Tag -Archive $RequestedArchive)) {
        Write-Host "Version $Tag has no published binaries for windows/$TargetArch." -ForegroundColor Yellow
        Write-Host "Falling back to the latest available release automatically..."
        $Tag = $null
    }
}

if (-not $Tag) {
    Write-Host "Determining latest release for $Repo..."
    $Tag = Get-LatestTag

    if (-not $Tag) {
        foreach ($Fallback in $FallbackTags) {
            Write-Host "Could not query GitHub API; trying known-good release $Fallback..."
            $ArchiveName = "agent-ledger_${Fallback}_windows_${TargetArch}.zip"
            if (Test-AssetExists -Tag $Fallback -Archive $ArchiveName) {
                $Tag = $Fallback
                break
            }
        }
    }

    if (-not $Tag) {
        Write-Error "Unable to determine an available release. Check https://github.com/$Repo/releases"
        exit 1
    }
    Write-Host "Using version: $Tag"
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
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing -Headers @{ "User-Agent" = "agent-ledger-installer" } -ErrorAction Stop
    } catch {
        Write-Host ""
        Write-Host "Download failed for $Tag ($($_.Exception.Message))." -ForegroundColor Yellow
        Write-Host "Verify the release exists at: https://github.com/$Repo/releases"
        exit 1
    }

    Write-Host "Extracting binaries..."
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    $AgentLedgerExe = Join-Path $TempDir "agent-ledger.exe"

    if (-not (Test-Path $AgentLedgerExe)) {
        Write-Error "Release archive did not contain agent-ledger.exe."
        exit 1
    }

    Write-Host "Installing executable to $InstallDir..."
    Copy-Item -Path $AgentLedgerExe -Destination (Join-Path $InstallDir "agent-ledger.exe") -Force

    # Verify installation
    $InstalledAgentLedger = Join-Path $InstallDir "agent-ledger.exe"

    if (-not (Test-Path $InstalledAgentLedger)) {
        Write-Error "Installation verification failed in $InstallDir."
        exit 1
    }

    Write-Host ""
    Write-Host "Successfully installed Agent Ledger ($Tag)!" -ForegroundColor Green
    Write-Host "  - Binary: $InstalledAgentLedger"
    Write-Host ""

    # Ensure InstallDir is on the User PATH (persists across sessions and reboots)
    $UserPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
    $UserPathEntries = @($UserPath -split ";" | Where-Object { $_ })
    if ($UserPathEntries -notcontains $InstallDir) {
        Write-Host "Adding $InstallDir to your User PATH..."
        $NewUserPath = (($UserPathEntries + $InstallDir) -join ";")
        [System.Environment]::SetEnvironmentVariable("Path", $NewUserPath, "User")

        # Broadcast environment change to all running processes so they pick up the new PATH immediately
        try {
            Add-Type -Name WinAPI -Namespace AgentLedger -MemberDefinition @"
using System;
using System.Runtime.InteropServices;
public class NativeMethods {
    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd, uint Msg, IntPtr wParam, string lParam,
        uint fuFlags, uint uTimeout, out IntPtr lpdwResult);
}
"@
            $HWND_BROADCAST = 0xFFFF
            $WM_SETTINGCHANGE = 0x1A
            $SMTO_ABORTIFHUNG = 0x0002
            $null = [AgentLedger.NativeMethods]::SendMessageTimeout(
                $HWND_BROADCAST, $WM_SETTINGCHANGE, [IntPtr]::Zero, "Environment",
                $SMTO_ABORTIFHUNG, 5000, [ref]0)
            Write-Host "Broadcasted PATH change to running applications." -ForegroundColor Cyan
        } catch {
            Write-Host "Note: Could not broadcast PATH change to running apps (may need admin)." -ForegroundColor Yellow
        }
    }

    # Make commands available in the current session immediately
    $CurrentPathEntries = @($env:Path -split ";" | Where-Object { $_ })
    if ($CurrentPathEntries -notcontains $InstallDir) {
        $env:Path = (($CurrentPathEntries + $InstallDir) -join ";")
    }

    Write-Host "Verification:"
    Write-Host "  Run 'agent-ledger --help' to get started."
    Write-Host "  Run 'agent-ledger mcp' to start the MCP server."
    if ($Host.Name -eq "ConsoleHost") {
        Write-Host ""
        Write-Host "Note: restart your terminal for PATH changes to apply in new sessions."
    }
} finally {
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
