param(
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",
  [string]$ClientSlug = "",
  [string]$EnvironmentId = "",
  [string]$CreatedBy = "backup.ps1",
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
  if ($null -ne $Body) {
    $options.Body = ($Body | ConvertTo-Json -Depth 8)
  }
  return Invoke-RestMethod @options
}

$composeFile = Join-Path $PSScriptRoot "docker-compose.client.yml"
$envFile = Join-Path $PSScriptRoot ".env.client"

if (!(Test-Path $envFile)) {
  throw "Missing .env.client. Copy .env.client.example to .env.client and fill client values."
}

$envValues = @{}
Get-Content $envFile | ForEach-Object {
  if ($_ -match "^\s*#" -or $_ -notmatch "=") { return }
  $key, $value = $_ -split "=", 2
  $envValues[$key.Trim()] = $value.Trim()
}

if (!$ClientSlug -and $envValues["CLIENT_SLUG"]) {
  $ClientSlug = $envValues["CLIENT_SLUG"]
}

$backupRoot = $envValues["BACKUP_DIR"]
if (!$backupRoot) { $backupRoot = Join-Path $PSScriptRoot "backups" }
if (![System.IO.Path]::IsPathRooted($backupRoot)) {
  $backupRoot = Join-Path $PSScriptRoot $backupRoot
}

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$target = Join-Path $backupRoot $stamp
New-Item -ItemType Directory -Force -Path $target | Out-Null

$dbName = $envValues["POSTGRES_DB"]
$dbUser = $envValues["POSTGRES_USER"]

docker compose --env-file $envFile -f $composeFile exec -T db pg_dump -U $dbUser -d $dbName --clean --if-exists | Set-Content -Encoding UTF8 (Join-Path $target "database.sql")

$targetMount = (Resolve-Path -LiteralPath $target).Path -replace "\\", "/"
docker compose --env-file $envFile -f $composeFile run --rm --no-deps -v "${targetMount}:/backup" backend sh -c "cd /app && tar -czf /backup/uploads.tar.gz uploads"

$secretPattern = "(PASSWORD|SECRET|TOKEN|KEY|DATABASE_URL|JWT)"
$envSnapshot = Join-Path $target "env.snapshot"
Get-Content $envFile | ForEach-Object {
  if ($_ -match "^\s*#" -or $_ -notmatch "=") { return }
  $key, $value = $_ -split "=", 2
  if ($key -match $secretPattern) {
    "$key=<redacted>"
  } else {
    "$key=$value"
  }
} | Set-Content -LiteralPath $envSnapshot -Encoding UTF8

$sizeBytes = (Get-ChildItem -LiteralPath $target -Recurse -File | Measure-Object -Property Length -Sum).Sum
if ($null -eq $sizeBytes) { $sizeBytes = 0 }

$manifest = [ordered]@{
  clientSlug = $ClientSlug
  createdAt = (Get-Date).ToUniversalTime().ToString("o")
  createdBy = $CreatedBy
  status = "succeeded"
  databaseDump = "database.sql"
  uploadsArchive = "uploads.tar.gz"
  envSnapshot = "env.snapshot"
  sizeBytes = [int64]$sizeBytes
}
$manifest | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $target "manifest.json") -Encoding UTF8

if ($Register) {
  $client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
  if ($client) {
    if (!$EnvironmentId) {
      $environments = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$($client.item.id)/environments"
      if ($environments -and $environments.items -and $environments.items.Count -gt 0) {
        $EnvironmentId = $environments.items[0].id
      }
    }
    $backupPayload = @{
      environmentId = $EnvironmentId
      status = "succeeded"
      backupPath = $target
      sizeBytes = [int64]$sizeBytes
      createdBy = $CreatedBy
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/backups" $backupPayload | Out-Null
  }
}

Write-Output "Backup created: $target"
Write-Output "Backup size: $sizeBytes bytes"
