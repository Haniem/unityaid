param(
  [Parameter(Mandatory = $true)]
  [string]$ClientSlug,

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

$client = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$ClientSlug"
$environments = Invoke-ControlPlaneJson "Get" "/api/v1/clients/$($client.item.id)/environments"
if (!$environments.items -or $environments.items.Count -eq 0) {
  throw "No environments registered for client $ClientSlug"
}

foreach ($environment in $environments.items) {
  $backendUrl = [string]$environment.backend_url
  $status = "unknown"
  $responseMs = 0
  $details = @{ backendUrl = $backendUrl }

  if ($backendUrl) {
    $healthUrl = $backendUrl.TrimEnd("/") + "/health"
    try {
      $watch = [System.Diagnostics.Stopwatch]::StartNew()
      $response = Invoke-WebRequest -UseBasicParsing $healthUrl -TimeoutSec 10
      $watch.Stop()
      $responseMs = [int]$watch.ElapsedMilliseconds
      $status = if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500) { "healthy" } else { "degraded" }
      $details.statusCode = $response.StatusCode
    } catch {
      $status = "failed"
      $details.error = $_.Exception.Message
    }
  }

  $payload = @{
    environmentId = $environment.id
    component = "backend"
    status = $status
    responseMs = $responseMs
    details = ($details | ConvertTo-Json -Compress)
  }
  Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/health-checks" $payload | Out-Null

  if ($status -ne "healthy") {
    $alertPayload = @{
      environmentId = $environment.id
      severity = "critical"
      status = "open"
      title = "Backend healthcheck failed"
      message = "Backend check for $ClientSlug returned $status."
    }
    Invoke-ControlPlaneJson "Post" "/api/v1/clients/$($client.item.id)/alerts" $alertPayload | Out-Null
  }

  Write-Output "$ClientSlug/$($environment.name): $status ${responseMs}ms"
}
