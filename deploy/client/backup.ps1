param(
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",
  [string]$ClientSlug = "",
  [string]$EnvironmentId = "",
  [string]$CreatedBy = "backup.ps1",
  [string]$EncryptionKey = "",
  [switch]$Encrypt,
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

function Protect-FileAes([string]$Path, [string]$KeyMaterial) {
  $plainBytes = [System.IO.File]::ReadAllBytes($Path)
  $sha = [System.Security.Cryptography.SHA256]::Create()
  $keyBytes = $sha.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($KeyMaterial))
  $aes = [System.Security.Cryptography.Aes]::Create()
  $aes.Key = $keyBytes
  $aes.GenerateIV()
  $aes.Mode = [System.Security.Cryptography.CipherMode]::CBC
  $aes.Padding = [System.Security.Cryptography.PaddingMode]::PKCS7
  $encryptor = $aes.CreateEncryptor()
  $cipherBytes = $encryptor.TransformFinalBlock($plainBytes, 0, $plainBytes.Length)
  $outPath = "$Path.enc"
  [System.IO.File]::WriteAllBytes($outPath, $aes.IV + $cipherBytes)
  $encryptor.Dispose()
  $aes.Dispose()
  $sha.Dispose()
  Remove-Item -LiteralPath $Path -Force
  return $outPath
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

$databaseFile = "database.sql"
$uploadsFile = "uploads.tar.gz"
$encrypted = $false
if ($Encrypt) {
  if (!$EncryptionKey) {
    $EncryptionKey = $envValues["BACKUP_ENCRYPTION_KEY"]
  }
  if (!$EncryptionKey) {
    throw "EncryptionKey or BACKUP_ENCRYPTION_KEY is required when -Encrypt is used."
  }
  Protect-FileAes (Join-Path $target $databaseFile) $EncryptionKey | Out-Null
  Protect-FileAes (Join-Path $target $uploadsFile) $EncryptionKey | Out-Null
  $databaseFile = "database.sql.enc"
  $uploadsFile = "uploads.tar.gz.enc"
  $encrypted = $true
}

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
  encrypted = $encrypted
  encryption = if ($encrypted) { "aes-256-cbc-sha256-key" } else { "none" }
  databaseDump = $databaseFile
  uploadsArchive = $uploadsFile
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
