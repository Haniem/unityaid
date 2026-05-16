$ErrorActionPreference = "Stop"

$composeFile = Join-Path $PSScriptRoot "docker-compose.client.yml"
$envFile = Join-Path $PSScriptRoot ".env.client"

if (!(Test-Path $envFile)) {
  throw "Missing .env.client. Copy .env.client.example to .env.client and fill client values."
}

docker compose --env-file $envFile -f $composeFile --profile seed run --rm seed
