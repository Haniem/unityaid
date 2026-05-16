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
  if (!$found) { $updated += $line }
  Set-Content -LiteralPath $Path -Value $updated -Encoding UTF8
}

$scriptRoot = $PSScriptRoot
$clientPath = Join-Path (Join-Path $scriptRoot $ClientsRoot) $ClientSlug
$envPath = Join-Path $clientPath ".env.client"
if (!(Test-Path -LiteralPath $envPath)) {
  throw "Client env file not found: $envPath"
}

$config = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug/effective-config"
$configPath = Join-Path $clientPath "control-plane-config.json"
$config | ConvertTo-Json -Depth 12 | Set-Content -LiteralPath $configPath -Encoding UTF8

$features = @()
if ($config.plan -and $config.plan.features) {
  $features += $config.plan.features
}
if ($config.client -and $config.client.modules) {
  $features += $config.client.modules
}
foreach ($flag in $config.featureFlags) {
  if ($flag.is_enabled) {
    $features += $flag.code
  }
}
$featuresValue = ($features | Where-Object { $_ } | Select-Object -Unique) -join ","
$limitsValue = if ($config.plan -and $config.plan.limits) { ($config.plan.limits | ConvertTo-Json -Compress -Depth 8) } else { "{}" }

Set-EnvValue $envPath "CONTROL_PLANE_CONFIG_PATH" "./control-plane-config.json"
Set-EnvValue $envPath "ENABLED_MODULES" $featuresValue
Set-EnvValue $envPath "PLAN_LIMITS_JSON" $limitsValue

Write-Output "Client config synced: $configPath"
Write-Output "Enabled modules: $featuresValue"
