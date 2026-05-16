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
- `GET /api/v1/clients/{id}/effective-config`
- `GET /api/v1/clients/{id}/feature-flags`
- `POST /api/v1/clients/{id}/feature-flags`
- `GET /api/v1/clients/{id}/security-events`
- `POST /api/v1/clients/{id}/security-events`
- `GET /api/v1/clients/{id}/branding`
- `PUT /api/v1/clients/{id}/branding`
- `GET /api/v1/clients/{id}/environments`
- `POST /api/v1/clients/{id}/environments`
- `GET /api/v1/clients/{id}/domains`
- `POST /api/v1/clients/{id}/domains`
- `GET /api/v1/clients/{id}/versions`
- `POST /api/v1/clients/{id}/versions`
- `GET /api/v1/clients/{id}/maintenance-windows`
- `POST /api/v1/clients/{id}/maintenance-windows`
- `GET /api/v1/clients/{id}/deployments`
- `POST /api/v1/clients/{id}/deployments`
- `GET /api/v1/clients/{id}/backups`
- `POST /api/v1/clients/{id}/backups`
- `POST /api/v1/clients/{id}/backups/{backupId}/restore`
- `GET /api/v1/clients/{id}/migrations`
- `POST /api/v1/clients/{id}/migrations`
- `GET /api/v1/clients/{id}/health-checks`
- `POST /api/v1/clients/{id}/health-checks`
- `GET /api/v1/clients/{id}/alerts`
- `POST /api/v1/clients/{id}/alerts`

## Provision a client stack

`provision-client.ps1` generates a standalone client folder from `deploy/client`, creates `.env.client`, optionally starts the stack, and optionally registers the client in control plane.

Example without starting containers:

```powershell
.\provision-client.ps1 `
  -ClientSlug dobrye-ruki `
  -ClientName "Добрые руки" `
  -OwnerEmail admin@dobrye-ruki.example `
  -AdminPassword "ChangeMe123!" `
  -PrimaryDomain dobrye-ruki.unityaid.example `
  -ControlPlaneApiKey "change-me-control-plane-key" `
  -Register
```

Example with immediate compose startup:

```powershell
.\provision-client.ps1 `
  -ClientSlug demo-local `
  -ClientName "Demo Local" `
  -OwnerEmail admin@example.org `
  -AdminPassword "ChangeMe123!" `
  -FrontendPort 8188 `
  -BackendPort 8180 `
  -PostgresPort 55433 `
  -ControlPlaneApiKey "change-me-control-plane-key" `
  -Register `
  -Start
```

When `-PrimaryDomain` is provided with `-Register`, the provisioner also registers the primary domain and the initial version record in control plane.

## Generate routing for a client domain

`generate-routing.ps1` creates a Caddy-based routing bundle for a client domain. Caddy obtains and renews HTTPS certificates automatically when the domain points to the host.

```powershell
.\generate-routing.ps1 `
  -ClientSlug dobrye-ruki `
  -Domain volunteers.dobrye-ruki.example `
  -FrontendTarget localhost:8188 `
  -ControlPlaneApiKey "change-me-control-plane-key" `
  -Register
```

The generated files are placed into `deploy/routing/{clientSlug}`:

- `Caddyfile`;
- `docker-compose.routing.yml`.

## Update a client stack

`update-client.ps1` records a target version, validates compose in dry-run mode, or rebuilds and restarts the client services.

```powershell
.\update-client.ps1 `
  -ClientSlug dobrye-ruki `
  -AppVersion "2026.05.17" `
  -ReleaseChannel stable `
  -ControlPlaneApiKey "change-me-control-plane-key" `
  -Register `
  -DryRun
```

Without `-DryRun`, the script creates a pre-update backup, rebuilds backend/frontend images, runs migration/bootstrap jobs, starts services, and writes a deployment entry with the final status.

To skip the automatic pre-update backup explicitly:

```powershell
.\update-client.ps1 `
  -ClientSlug dobrye-ruki `
  -AppVersion "2026.05.17-hotfix.1" `
  -SkipPreUpdateBackup
```

## Backup registry

Client stacks can register backup results in control plane:

```powershell
cd ..\clients\dobrye-ruki
.\backup.ps1 `
  -ControlPlaneApiKey "change-me-control-plane-key" `
  -Register
```

The backup entry stores status, size, path, creator, environment and restore timestamp.

## Feature flags and limits

Control plane exposes an effective client config that combines plan limits, plan features, client modules and manual feature flags:

```powershell
.\sync-client-config.ps1 `
  -ClientSlug dobrye-ruki `
  -ControlPlaneApiKey "change-me-control-plane-key"
```

The script writes `control-plane-config.json` into the client stack and updates:

- `CONTROL_PLANE_CONFIG_PATH`;
- `ENABLED_MODULES`;
- `PLAN_LIMITS_JSON`.

Feature flag example:

```powershell
$client = Invoke-RestMethod http://localhost:8090/api/v1/clients/dobrye-ruki -Headers @{ "X-Control-Plane-Key" = "change-me-control-plane-key" }
Invoke-RestMethod `
  "http://localhost:8090/api/v1/clients/$($client.item.id)/feature-flags" `
  -Method Post `
  -Headers @{ "X-Control-Plane-Key" = "change-me-control-plane-key" } `
  -ContentType "application/json" `
  -Body '{"code":"qr_checkin","isEnabled":true,"config":"{}"}'
```

## Monitoring

`check-client-health.ps1` checks registered client environments and writes health-check records to control plane. Failed checks create open alerts.

```powershell
.\check-client-health.ps1 `
  -ClientSlug dobrye-ruki `
  -ControlPlaneApiKey "change-me-control-plane-key"
```

## Security operations

Rotate per-client secrets and write a technical audit event:

```powershell
.\rotate-client-secret.ps1 `
  -ClientSlug dobrye-ruki `
  -SecretName JWT_SECRET `
  -ControlPlaneApiKey "change-me-control-plane-key"
```

Client backup supports AES encryption:

```powershell
cd ..\clients\dobrye-ruki
.\backup.ps1 -Encrypt -ControlPlaneApiKey "change-me-control-plane-key" -Register
```

## Branding

Branding is stored in control plane and can be synchronized into a client stack:

```powershell
.\sync-client-branding.ps1 `
  -ClientSlug dobrye-ruki `
  -ControlPlaneApiKey "change-me-control-plane-key"
```
