param(
  [Parameter(Mandatory = $true)]
  [string]$BackupPath
)

$ErrorActionPreference = "Stop"

$composeFile = Join-Path $PSScriptRoot "docker-compose.client.yml"
$envFile = Join-Path $PSScriptRoot ".env.client"

if (!(Test-Path $envFile)) {
  throw "Missing .env.client. Copy .env.client.example to .env.client and fill client values."
}
if (!(Test-Path $BackupPath)) {
  throw "Backup path not found: $BackupPath"
}

$databaseDump = Join-Path $BackupPath "database.sql"
$uploadsArchive = Join-Path $BackupPath "uploads.tar.gz"

if (!(Test-Path $databaseDump)) {
  throw "database.sql not found in backup path."
}

$envValues = @{}
Get-Content $envFile | ForEach-Object {
  if ($_ -match "^\s*#" -or $_ -notmatch "=") { return }
  $key, $value = $_ -split "=", 2
  $envValues[$key.Trim()] = $value.Trim()
}

$dbName = $envValues["POSTGRES_DB"]
$dbUser = $envValues["POSTGRES_USER"]

Get-Content $databaseDump | docker compose --env-file $envFile -f $composeFile exec -T db psql -U $dbUser -d $dbName

if (Test-Path $uploadsArchive) {
  Get-Content -Encoding Byte $uploadsArchive | docker compose --env-file $envFile -f $composeFile run --rm --no-deps backend sh -c "cd /app && rm -rf uploads && tar -xzf -"
}

Write-Output "Backup restored: $BackupPath"
