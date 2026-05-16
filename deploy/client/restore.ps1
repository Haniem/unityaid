param(
  [Parameter(Mandatory = $true)]
  [string]$BackupPath,
  [string]$EncryptionKey = ""
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
$manifestPath = Join-Path $BackupPath "manifest.json"

$envValues = @{}
Get-Content $envFile | ForEach-Object {
  if ($_ -match "^\s*#" -or $_ -notmatch "=") { return }
  $key, $value = $_ -split "=", 2
  $envValues[$key.Trim()] = $value.Trim()
}

function Unprotect-FileAes([string]$Path, [string]$KeyMaterial, [string]$OutputPath) {
  $bytes = [System.IO.File]::ReadAllBytes($Path)
  $iv = New-Object byte[] 16
  [Array]::Copy($bytes, 0, $iv, 0, 16)
  $cipher = New-Object byte[] ($bytes.Length - 16)
  [Array]::Copy($bytes, 16, $cipher, 0, $cipher.Length)
  $sha = [System.Security.Cryptography.SHA256]::Create()
  $keyBytes = $sha.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($KeyMaterial))
  $aes = [System.Security.Cryptography.Aes]::Create()
  $aes.Key = $keyBytes
  $aes.IV = $iv
  $aes.Mode = [System.Security.Cryptography.CipherMode]::CBC
  $aes.Padding = [System.Security.Cryptography.PaddingMode]::PKCS7
  $decryptor = $aes.CreateDecryptor()
  $plain = $decryptor.TransformFinalBlock($cipher, 0, $cipher.Length)
  [System.IO.File]::WriteAllBytes($OutputPath, $plain)
  $decryptor.Dispose()
  $aes.Dispose()
  $sha.Dispose()
}

if (Test-Path $manifestPath) {
  $manifest = Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
  if ($manifest.encrypted) {
    if (!$EncryptionKey) {
      $EncryptionKey = $envValues["BACKUP_ENCRYPTION_KEY"]
    }
    if (!$EncryptionKey) {
      throw "EncryptionKey or BACKUP_ENCRYPTION_KEY is required to restore encrypted backup."
    }
    Unprotect-FileAes (Join-Path $BackupPath $manifest.databaseDump) $EncryptionKey $databaseDump
    Unprotect-FileAes (Join-Path $BackupPath $manifest.uploadsArchive) $EncryptionKey $uploadsArchive
  }
} else {
  Write-Warning "manifest.json not found in backup path. Continuing with legacy backup format."
}

if (!(Test-Path $databaseDump)) {
  throw "database.sql not found in backup path."
}

$dbName = $envValues["POSTGRES_DB"]
$dbUser = $envValues["POSTGRES_USER"]

Get-Content $databaseDump | docker compose --env-file $envFile -f $composeFile exec -T db psql -U $dbUser -d $dbName

if (Test-Path $uploadsArchive) {
  $backupMount = (Resolve-Path -LiteralPath $BackupPath).Path -replace "\\", "/"
  docker compose --env-file $envFile -f $composeFile run --rm --no-deps -v "${backupMount}:/restore" backend sh -c "cd /app && rm -rf uploads && tar -xzf /restore/uploads.tar.gz"
}

Write-Output "Backup restored: $BackupPath"
