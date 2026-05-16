param(
  [Parameter(Mandatory = $true)]
  [string]$ClientSlug,

  [string]$ClientsRoot = "../clients",
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [Parameter(Mandatory = $true)]
  [string]$ControlPlaneApiKey
)

$ErrorActionPreference = "Stop"

function Invoke-ControlPlaneJson($Method, $Path, $Body = $null) {
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

$scriptRoot = $PSScriptRoot
$clientPath = Join-Path (Join-Path $scriptRoot $ClientsRoot) $ClientSlug
$envPath = Join-Path $clientPath ".env.client"
if (!(Test-Path -LiteralPath $envPath)) {
  throw "Client env file not found: $envPath"
}

$client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
$branding = (Invoke-ControlPlaneJson "Get" "/api/v1/clients/$($client.item.id)/branding").item
$brandingPath = Join-Path $clientPath "branding.json"
$branding | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $brandingPath -Encoding UTF8

Set-EnvValue $envPath "CLIENT_DISPLAY_NAME" $branding.display_name
Set-EnvValue $envPath "BRAND_PRIMARY_COLOR" $branding.primary_color
Set-EnvValue $envPath "BRAND_SECONDARY_COLOR" $branding.secondary_color
Set-EnvValue $envPath "CLIENT_TIMEZONE" $branding.timezone
Set-EnvValue $envPath "CLIENT_LOCALE" $branding.locale
Set-EnvValue $envPath "CLIENT_BRANDING_PATH" "./branding.json"

Write-Output "Client branding synced: $brandingPath"
