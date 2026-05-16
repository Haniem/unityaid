param(
  [Parameter(Mandatory = $true)]
  [string]$ArchivePath,

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

if (!(Test-Path -LiteralPath $ArchivePath)) {
  throw "Archive path not found: $ArchivePath"
}

$manifestPath = Join-Path $ArchivePath "migration-manifest.json"
if (Test-Path -LiteralPath $manifestPath) {
  $manifest = Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
  if (!$ClientSlug -and $manifest.clientSlug) {
    $ClientSlug = $manifest.clientSlug
  }
}

$backupPath = Join-Path $ArchivePath "backup"
if (!(Test-Path -LiteralPath $backupPath)) {
  throw "Backup folder not found in archive."
}
& (Join-Path $PSScriptRoot "restore.ps1") -BackupPath $backupPath

if ($Register -and $ClientSlug) {
  $client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
  if ($client) {
    $payload = @{
      kind = "import"
      status = "succeeded"
      source = $ArchivePath
      target = $ClientSlug
      archivePath = $ArchivePath
      log = "Client import restored"
      createdBy = "import-client.ps1"
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/migrations" $payload | Out-Null
  }
}

Write-Output "Client import restored: $ArchivePath"
