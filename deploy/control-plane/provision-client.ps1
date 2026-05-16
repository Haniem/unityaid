param(
  [Parameter(Mandatory = $true)]
  [string]$ClientSlug,

  [Parameter(Mandatory = $true)]
  [string]$ClientName,

  [Parameter(Mandatory = $true)]
  [string]$OwnerEmail,

  [string]$PrimaryDomain = "",
  [string]$PlanCode = "start",
  [string[]]$Modules = @("events", "tasks", "time_entries", "certificates"),

  [string]$AdminEmail = "",
  [Parameter(Mandatory = $true)]
  [string]$AdminPassword,
  [string]$AdminFirstName = "Admin",
  [string]$AdminLastName = "UnityAid",
  [string]$OrganizationName = "",
  [string]$OrganizationSlug = "",

  [string]$AppVersion = "local",
  [int]$FrontendPort = 8088,
  [int]$BackendPort = 8080,
  [int]$PostgresPort = 5432,
  [int]$RedisPort = 6379,

  [string]$ClientsRoot = "../clients",
  [string]$TemplateRoot = "../client",
  [string]$ControlPlaneUrl = "http://localhost:8090",
  [string]$ControlPlaneApiKey = "",

  [switch]$WithRedis,
  [switch]$WithSeeds,
  [switch]$Start,
  [switch]$Register
)

$ErrorActionPreference = "Stop"

function Normalize-Slug([string]$value) {
  $normalized = $value.Trim().ToLowerInvariant() -replace "[^a-z0-9-]", "-"
  $normalized = $normalized -replace "-+", "-"
  return $normalized.Trim("-")
}

function New-RandomSecret([int]$bytes = 32) {
  $buffer = New-Object byte[] $bytes
  $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try {
    $rng.GetBytes($buffer)
    return [Convert]::ToBase64String($buffer)
  } finally {
    $rng.Dispose()
  }
}

function Invoke-ControlPlaneJson($Method, $Path, $Body = $null) {
  if (!$ControlPlaneApiKey) {
    Write-Warning "CONTROL_PLANE_API_KEY is empty. Skipping control plane registration call: $Method $Path"
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
if (!$ClientSlug) {
  throw "ClientSlug is empty after normalization."
}
if (!$AdminEmail) {
  $AdminEmail = $OwnerEmail
}
if (!$OrganizationName) {
  $OrganizationName = $ClientName
}
if (!$OrganizationSlug) {
  $OrganizationSlug = $ClientSlug
}

$scriptRoot = $PSScriptRoot
$templatePath = Resolve-Path (Join-Path $scriptRoot $TemplateRoot)
$clientsRootPath = Join-Path $scriptRoot $ClientsRoot
$clientPath = Join-Path $clientsRootPath $ClientSlug

if (Test-Path $clientPath) {
  throw "Client folder already exists: $clientPath"
}

New-Item -ItemType Directory -Force -Path $clientsRootPath | Out-Null
New-Item -ItemType Directory -Force -Path $clientPath | Out-Null
Copy-Item -Path (Join-Path $templatePath "*") -Destination $clientPath -Recurse -Force

$publicBaseUrl = if ($PrimaryDomain) { "https://$PrimaryDomain" } else { "http://localhost:$FrontendPort" }
$corsOrigin = $publicBaseUrl
$composeProject = "unityaid_$($ClientSlug -replace '-', '_')"
$profiles = @()
if ($WithRedis) { $profiles += "redis" }
$profilesValue = $profiles -join ","

$envContent = @"
CLIENT_SLUG=$ClientSlug
COMPOSE_PROJECT_NAME=$composeProject
COMPOSE_PROFILES=$profilesValue

APP_ENV=production
PUBLIC_BASE_URL=$publicBaseUrl
FRONTEND_URL=$publicBaseUrl
BACKEND_URL=$publicBaseUrl/api/v1

POSTGRES_USER=unityaid
POSTGRES_PASSWORD=$(New-RandomSecret 24)
POSTGRES_DB=unityaid
POSTGRES_PORT=$PostgresPort
REDIS_PORT=$RedisPort

BACKEND_PORT=$BackendPort
FRONTEND_PORT=$FrontendPort

CORS_ALLOWED_ORIGINS=$corsOrigin
VITE_API_BASE_URL=/api/v1

JWT_SECRET=$(New-RandomSecret 48)
RUN_SEEDS=false
UPLOADS_DIR=/app/uploads

BOOTSTRAP_ADMIN_EMAIL=$AdminEmail
BOOTSTRAP_ADMIN_PASSWORD=$AdminPassword
BOOTSTRAP_ADMIN_FIRST_NAME=$AdminFirstName
BOOTSTRAP_ADMIN_LAST_NAME=$AdminLastName
BOOTSTRAP_ADMIN_PATRONYMIC=
BOOTSTRAP_ORGANIZATION_NAME=$OrganizationName
BOOTSTRAP_ORGANIZATION_SLUG=$OrganizationSlug

BACKUP_DIR=./backups
"@

$envPath = Join-Path $clientPath ".env.client"
Set-Content -LiteralPath $envPath -Value $envContent -Encoding UTF8

if ($Start) {
  Push-Location $clientPath
  try {
    & .\start.ps1
    if ($WithSeeds) {
      & .\seed.ps1
    }
  } finally {
    Pop-Location
  }
}

$clientId = $null
if ($Register) {
  $clientPayload = @{
    slug = $ClientSlug
    name = $ClientName
    ownerEmail = $OwnerEmail
    planCode = $PlanCode
    primaryDomain = $PrimaryDomain
    modules = $Modules
  }

  try {
    $clientResponse = Invoke-ControlPlaneJson "Post" "/api/v1/clients" $clientPayload
    $clientId = $clientResponse.item.id
  } catch {
    Write-Warning "Client create failed, trying to load existing client by slug. $($_.Exception.Message)"
    $clientResponse = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
    if ($clientResponse) {
      $clientId = $clientResponse.item.id
    }
  }

  if ($clientId) {
    $environmentStatus = if ($Start) { "running" } else { "planned" }
    $environmentPayload = @{
      name = "production"
      kind = "compose"
      status = $environmentStatus
      stackPath = (Resolve-Path $clientPath).Path
      appVersion = $AppVersion
      frontendUrl = $publicBaseUrl
      backendUrl = "$publicBaseUrl/api/v1"
    }
    $environmentResponse = Invoke-ControlPlaneJson "Post" "/api/v1/clients/$clientId/environments" $environmentPayload

    $deploymentPayload = @{
      environmentId = $environmentResponse.item.id
      version = $AppVersion
      status = if ($Start) { "succeeded" } else { "planned" }
      triggeredBy = "provision-client.ps1"
      log = "Client stack generated at $clientPath"
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$clientId/deployments" $deploymentPayload | Out-Null
  }
}

Write-Output "Client stack created: $clientPath"
Write-Output "Frontend URL: $publicBaseUrl"
Write-Output "Admin email: $AdminEmail"
if ($Register -and $clientId) {
  Write-Output "Control plane client id: $clientId"
}
