param(
  [string]$ExportDir = "",
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",
  [string]$ClientSlug = "",
  [switch]$Register
)

$ErrorActionPreference = "Stop"

function Invoke-ControlPlaneJson($Method, $Path, $Body = $null) {
  if (!$ControlPlaneApiKey) {
    Write-Warning "CONTROL_PLANE_API_KEY is empty. Skipping control plane call: $Method $Path"
    return $null
  }
  $headers = @{ "X-Control-Plane-Key" = $ControlPlaneApiKey }
  $uri = "$($ControlPlaneUrl.TrimEnd('/'))$Path"
  $options = @{
    Method = $Method
    Uri = $uri
    Headers = $headers
    ContentType = "application/json"
  }
  if ($null -ne $Body) { $options.Body = ($Body | ConvertTo-Json -Depth 8) }
  return Invoke-RestMethod @options
}

$envFile = Join-Path $PSScriptRoot ".env.client"
if (!(Test-Path -LiteralPath $envFile)) {
  throw "Missing .env.client."
}

$envValues = @{}
Get-Content -LiteralPath $envFile | ForEach-Object {
  if ($_ -match "^\s*#" -or $_ -notmatch "=") { return }
  $key, $value = $_ -split "=", 2
  $envValues[$key.Trim()] = $value.Trim()
}
if (!$ClientSlug -and $envValues["CLIENT_SLUG"]) {
  $ClientSlug = $envValues["CLIENT_SLUG"]
}
if (!$ClientSlug) { throw "ClientSlug is required." }

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
if (!$ExportDir) {
  $ExportDir = Join-Path $PSScriptRoot "exports"
}
if (![System.IO.Path]::IsPathRooted($ExportDir)) {
  $ExportDir = Join-Path $PSScriptRoot $ExportDir
}
$target = Join-Path $ExportDir "$($ClientSlug)_$stamp"
New-Item -ItemType Directory -Force -Path $target | Out-Null

& (Join-Path $PSScriptRoot "backup.ps1") -CreatedBy "export-client.ps1"
$latestBackup = Get-ChildItem -LiteralPath (Join-Path $PSScriptRoot "backups") -Directory | Sort-Object LastWriteTime -Descending | Select-Object -First 1
if ($latestBackup) {
  Copy-Item -LiteralPath $latestBackup.FullName -Destination (Join-Path $target "backup") -Recurse -Force
}
if (Test-Path -LiteralPath (Join-Path $PSScriptRoot "control-plane-config.json")) {
  Copy-Item -LiteralPath (Join-Path $PSScriptRoot "control-plane-config.json") -Destination (Join-Path $target "control-plane-config.json") -Force
}

$manifest = [ordered]@{
  clientSlug = $ClientSlug
  exportedAt = (Get-Date).ToUniversalTime().ToString("o")
  source = "client-stack"
  backupPath = "backup"
}
$manifest | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $target "migration-manifest.json") -Encoding UTF8

if ($Register) {
  $client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
  if ($client) {
    $payload = @{
      kind = "export"
      status = "succeeded"
      source = $ClientSlug
      target = ""
      archivePath = $target
      log = "Client export prepared"
      createdBy = "export-client.ps1"
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/migrations" $payload | Out-Null
  }
}

Write-Output "Client export created: $target"
