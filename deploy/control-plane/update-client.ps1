param(
  [Parameter(Mandatory = $true)]
  [string]$ClientSlug,

  [Parameter(Mandatory = $true)]
  [string]$AppVersion,

  [string]$BackendImage = "",
  [string]$FrontendImage = "",
  [string]$DBSchemaVersion = "",
  [string]$ReleaseChannel = "stable",
  [string]$ClientsRoot = "../clients",
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",
  [string]$TriggeredBy = "update-client.ps1",
  [switch]$DryRun,
  [switch]$SkipPreUpdateBackup,
  [switch]$Register
)

$ErrorActionPreference = "Stop"

function Normalize-Slug([string]$value) {
  $normalized = $value.Trim().ToLowerInvariant() -replace "[^a-z0-9-]", "-"
  $normalized = $normalized -replace "-+", "-"
  return $normalized.Trim("-")
}

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
  if ($null -ne $Body) {
    $options.Body = ($Body | ConvertTo-Json -Depth 8)
  }
  return Invoke-RestMethod @options
}

function Set-EnvValue([string]$Path, [string]$Key, [string]$Value) {
  $lines = Get-Content -LiteralPath $Path
  $escapedKey = [regex]::Escape($Key)
  $line = "$Key=$Value"
  $found = $false
  $updated = $lines | ForEach-Object {
    if ($_ -match "^$escapedKey=") {
      $found = $true
      $line
    } else {
      $_
    }
  }
  if (!$found) {
    $updated += $line
  }
  Set-Content -LiteralPath $Path -Value $updated -Encoding UTF8
}

$ClientSlug = Normalize-Slug $ClientSlug
if (!$ClientSlug) { throw "ClientSlug is empty after normalization." }
if (!$AppVersion.Trim()) { throw "AppVersion is required." }

$scriptRoot = $PSScriptRoot
$clientPath = Join-Path (Join-Path $scriptRoot $ClientsRoot) $ClientSlug
$envPath = Join-Path $clientPath ".env.client"
if (!(Test-Path -LiteralPath $clientPath)) { throw "Client folder not found: $clientPath" }
if (!(Test-Path -LiteralPath $envPath)) { throw "Client env file not found: $envPath" }

Set-EnvValue $envPath "APP_VERSION" $AppVersion
if ($BackendImage) { Set-EnvValue $envPath "BACKEND_IMAGE" $BackendImage }
if ($FrontendImage) { Set-EnvValue $envPath "FRONTEND_IMAGE" $FrontendImage }
if ($DBSchemaVersion) { Set-EnvValue $envPath "DB_SCHEMA_VERSION" $DBSchemaVersion }
Set-EnvValue $envPath "RELEASE_CHANNEL" $ReleaseChannel

$clientId = $null
$environmentId = ""
if ($Register) {
  $client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
  if ($client) {
    $clientId = $client.item.id
    $versionPayload = @{
      backendImage = $BackendImage
      frontendImage = $FrontendImage
      appVersion = $AppVersion
      dbSchemaVersion = $DBSchemaVersion
      releaseChannel = $ReleaseChannel
      status = if ($DryRun) { "dry_run" } else { "planned" }
      notes = "Prepared by $TriggeredBy"
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$clientId/versions" $versionPayload | Out-Null

    $environments = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$clientId/environments"
    if ($environments -and $environments.items -and $environments.items.Count -gt 0) {
      $environmentId = $environments.items[0].id
    }
  }
}

$status = "succeeded"
$log = "Updated client stack to $AppVersion"

Push-Location $clientPath
try {
  if ($DryRun) {
    docker compose --env-file .env.client -f docker-compose.client.yml config --quiet
    $status = "dry_run_succeeded"
    $log = "Dry-run config validation succeeded for $AppVersion"
  } else {
    if (!$SkipPreUpdateBackup -and (Test-Path -LiteralPath ".\backup.ps1")) {
      $backupArgs = @{
        CreatedBy = "pre-update:$TriggeredBy"
      }
      if ($Register) {
        $backupArgs.Register = $true
        $backupArgs.ControlPlaneUrl = $ControlPlaneUrl
        $backupArgs.ControlPlaneApiKey = $ControlPlaneApiKey
        $backupArgs.ClientSlug = $ClientSlug
        $backupArgs.EnvironmentId = $environmentId
      }
      & .\backup.ps1 @backupArgs
    }

    $profilesLine = Get-Content -LiteralPath $envPath | Where-Object { $_ -match "^COMPOSE_PROFILES=" } | Select-Object -First 1
    $profilesValue = if ($profilesLine) { $profilesLine.Split("=", 2)[1] } else { "" }
    $services = @("migrate", "bootstrap-admin", "backend", "frontend")
    if ($profilesValue -match "(^|,)worker(,|$)") {
      $services += "worker"
    }

    docker compose --env-file .env.client -f docker-compose.client.yml build backend frontend migrate bootstrap-admin
    if ($services -contains "worker") {
      docker compose --env-file .env.client -f docker-compose.client.yml build worker
    }
    docker compose --env-file .env.client -f docker-compose.client.yml up -d @services
    docker compose --env-file .env.client -f docker-compose.client.yml ps

    $healthUrl = ((Get-Content -LiteralPath $envPath | Where-Object { $_ -match "^PUBLIC_BASE_URL=" } | Select-Object -First 1) -replace "^PUBLIC_BASE_URL=", "").TrimEnd("/")
    if ($healthUrl -match "^http://localhost:|^http://127\.0\.0\.1:") {
      Invoke-WebRequest -UseBasicParsing "$healthUrl/api/v1/health" | Out-Null
    }
  }
} catch {
  $status = "failed"
  $log = $_.Exception.Message
  throw
} finally {
  Pop-Location
  if ($Register -and $clientId) {
    $deploymentPayload = @{
      environmentId = $environmentId
      version = $AppVersion
      status = $status
      triggeredBy = $TriggeredBy
      log = $log
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$clientId/deployments" $deploymentPayload | Out-Null
  }
}

Write-Output "Client update prepared: $clientPath"
Write-Output "Version: $AppVersion"
Write-Output "Status: $status"
