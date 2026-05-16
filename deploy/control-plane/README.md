# UnityAid control plane MVP

The control plane is the platform-side service for managing isolated client stacks.

It stores technical metadata only:

- clients;
- environments;
- domains and URLs;
- app versions;
- plans;
- deployments;
- backups;
- feature flags.

It does not store volunteer operational data. Client data remains inside each client stack.

## Start

```powershell
Copy-Item .env.example .env
docker compose --env-file .env up -d --build
```

Healthcheck:

```powershell
Invoke-WebRequest http://localhost:8090/health
```

## API authentication

Protected endpoints require one of:

```text
X-Control-Plane-Key: change-me-control-plane-key
Authorization: Bearer change-me-control-plane-key
```

## Example

```powershell
$headers = @{ "X-Control-Plane-Key" = "change-me-control-plane-key" }
$body = @{
  slug = "dobrye-ruki"
  name = "Добрые руки"
  ownerEmail = "admin@dobrye-ruki.example"
  planCode = "organization"
  primaryDomain = "dobrye-ruki.unityaid.example"
  modules = @("events", "tasks", "time_entries", "analytics", "certificates")
} | ConvertTo-Json

Invoke-WebRequest `
  -Method Post `
  -Uri http://localhost:8090/api/v1/clients `
  -Headers $headers `
  -ContentType "application/json" `
  -Body $body
```

## Endpoints

- `GET /health`
- `GET /api/v1/plans`
- `GET /api/v1/clients`
- `POST /api/v1/clients`
- `GET /api/v1/clients/{idOrSlug}`
- `PATCH /api/v1/clients/{id}`
- `GET /api/v1/clients/{id}/environments`
- `POST /api/v1/clients/{id}/environments`
- `GET /api/v1/clients/{id}/deployments`
- `POST /api/v1/clients/{id}/deployments`
- `GET /api/v1/clients/{id}/backups`
- `POST /api/v1/clients/{id}/backups`
