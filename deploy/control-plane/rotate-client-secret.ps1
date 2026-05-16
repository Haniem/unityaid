param(
  [Parameter(Mandatory = $true)]
  [string]$ClientSlug,

  [ValidateSet("JWT_SECRET", "BACKUP_ENCRYPTION_KEY")]
  [string]$SecretName = "JWT_SECRET",

  [string]$ClientsRoot = "../clients",
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",
  [string]$Actor = "rotate-client-secret.ps1"
)

$ErrorActionPreference = "Stop"

function New-RandomSecret([int]$bytes = 48) {
  $buffer = New-Object byte[] $bytes
  $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try {
    $rng.GetBytes($buffer)
    return [Convert]::ToBase64String($buffer)
  } finally {
    $rng.Dispose()
  }
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
  if (!$found) { $updated += $line }
  Set-Content -LiteralPath $Path -Value $updated -Encoding UTF8
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
  if ($null -ne $Body) { $options.Body = ($Body | ConvertTo-Json -Depth 8) }
  return Invoke-RestMethod @options
}

$scriptRoot = $PSScriptRoot
$clientPath = Join-Path (Join-Path $scriptRoot $ClientsRoot) $ClientSlug
$envPath = Join-Path $clientPath ".env.client"
if (!(Test-Path -LiteralPath $envPath)) {
  throw "Client env file not found: $envPath"
}

$newSecret = New-RandomSecret
Set-EnvValue $envPath $SecretName $newSecret

$client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
if ($client) {
  $payload = @{
    actor = $Actor
    action = "rotate_secret"
    target = $SecretName
    metadata = (@{ secretName = $SecretName; envPath = $envPath } | ConvertTo-Json -Compress)
  }
  Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/security-events" $payload | Out-Null
}

Write-Output "Secret rotated: $SecretName"
Write-Output "Client env updated: $envPath"
