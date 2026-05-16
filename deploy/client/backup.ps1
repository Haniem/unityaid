$ErrorActionPreference = "Stop"

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
docker compose --env-file $envFile -f $composeFile run --rm --no-deps backend sh -c "cd /app && tar -czf - uploads" > (Join-Path $target "uploads.tar.gz")

Copy-Item $envFile (Join-Path $target "env.snapshot") -Force
Write-Output "Backup created: $target"
