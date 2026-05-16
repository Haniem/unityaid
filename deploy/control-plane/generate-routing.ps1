param(
  [Parameter(Mandatory = $true)]
  [string]$ClientSlug,

  [Parameter(Mandatory = $true)]
  [string]$Domain,

  [string]$FrontendTarget = "",
  [string]$OutputRoot = "../routing",
  [string]$ClientsRoot = "../clients",
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",
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

$ClientSlug = Normalize-Slug $ClientSlug
$Domain = $Domain.Trim().ToLowerInvariant()
if (!$ClientSlug) { throw "ClientSlug is empty after normalization." }
if (!$Domain) { throw "Domain is required." }

$scriptRoot = $PSScriptRoot
$outputPath = Join-Path (Join-Path $scriptRoot $OutputRoot) $ClientSlug
New-Item -ItemType Directory -Force -Path $outputPath | Out-Null

if (!$FrontendTarget) {
  $clientEnvPath = Join-Path (Join-Path $scriptRoot $ClientsRoot) (Join-Path $ClientSlug ".env.client")
  if (Test-Path -LiteralPath $clientEnvPath) {
    $frontendPortLine = Get-Content -LiteralPath $clientEnvPath | Where-Object { $_ -match "^FRONTEND_PORT=" } | Select-Object -First 1
    $frontendPort = if ($frontendPortLine) { $frontendPortLine.Split("=", 2)[1].Trim() } else { "8088" }
    $FrontendTarget = "localhost:$frontendPort"
  } else {
    $FrontendTarget = "$ClientSlug-frontend:80"
  }
}

$caddyfile = @"
$Domain {
  encode gzip zstd
  reverse_proxy $FrontendTarget
}
"@

$caddyPath = Join-Path $outputPath "Caddyfile"
Set-Content -LiteralPath $caddyPath -Value $caddyfile -Encoding UTF8

$compose = @"
services:
  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
"@

$composePath = Join-Path $outputPath "docker-compose.routing.yml"
Set-Content -LiteralPath $composePath -Value $compose -Encoding UTF8

if ($Register) {
  $client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
  if ($client) {
    $payload = @{
      domain = $Domain
      kind = "custom"
      status = "planned"
      sslStatus = "planned"
      routeTarget = $FrontendTarget
      isPrimary = $true
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/domains" $payload | Out-Null
  }
}

Write-Output "Routing config created: $outputPath"
Write-Output "Domain: https://$Domain"
Write-Output "Target: $FrontendTarget"
