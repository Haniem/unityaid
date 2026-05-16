# UnityAid Control Plane API

Control plane MVP is a separate platform service for tracking isolated client stacks in the managed single-tenant model.

It keeps platform metadata, not volunteer data.

## Security

All `/api/v1/*` endpoints are protected with a platform API key.

Supported headers:

```text
X-Control-Plane-Key: <key>
Authorization: Bearer <key>
```

## Core resources

### Plans

`GET /api/v1/plans`

Returns available commercial plans and their feature/limit metadata.

### Clients

`GET /api/v1/clients?search=&limit=50&offset=0`

Returns client registry.

`POST /api/v1/clients`

Creates a client record.

```json
{
  "slug": "dobrye-ruki",
  "name": "Добрые руки",
  "ownerEmail": "admin@dobrye-ruki.example",
  "planCode": "organization",
  "primaryDomain": "dobrye-ruki.unityaid.example",
  "modules": ["events", "tasks", "time_entries", "analytics", "certificates"]
}
```

`GET /api/v1/clients/{idOrSlug}`

Returns one client by UUID or slug.

`PATCH /api/v1/clients/{id}`

Updates client metadata.

### Environments

`POST /api/v1/clients/{id}/environments`

Registers a client environment.

```json
{
  "name": "production",
  "kind": "compose",
  "status": "running",
  "stackPath": "deploy/clients/dobrye-ruki",
  "appVersion": "1.0.0",
  "frontendUrl": "https://dobrye-ruki.unityaid.example",
  "backendUrl": "https://dobrye-ruki.unityaid.example/api/v1"
}
```

`GET /api/v1/clients/{id}/environments`

Returns environments for a client.

### Deployments

`POST /api/v1/clients/{id}/deployments`

Registers deployment history.

`GET /api/v1/clients/{id}/deployments`

Returns deployment history.

### Backups

`POST /api/v1/clients/{id}/backups`

Registers backup metadata.

`GET /api/v1/clients/{id}/backups`

Returns backup history.
