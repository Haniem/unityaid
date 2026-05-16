# UnityAid client stack

This folder contains the first production-oriented client stack template for the managed single-tenant model.

## Quick start

1. Copy `.env.client.example` to `.env.client`.
2. Fill secrets, domains, ports and bootstrap admin values.
3. Start the stack:

```powershell
.\start.ps1
```

The stack creates:

- PostgreSQL database;
- migration job;
- bootstrap admin job;
- backend API;
- production frontend served by nginx;
- persistent uploads volume.

## Version fields

The stack keeps release metadata in `.env.client`:

- `APP_VERSION`;
- `RELEASE_CHANNEL`;
- `BACKEND_IMAGE`;
- `FRONTEND_IMAGE`;
- `DB_SCHEMA_VERSION`.

Control plane scripts use these fields to record planned and completed client updates.

## Plan limits and modules

The stack can receive commercial configuration from control plane:

- `CONTROL_PLANE_CONFIG_PATH`;
- `ENABLED_MODULES`;
- `PLAN_LIMITS_JSON`.

Run `deploy/control-plane/sync-client-config.ps1` to refresh these values from the client's plan, modules and feature flags.

## Branding

Branding values are stored in `.env.client` and can be refreshed from control plane:

- `CLIENT_DISPLAY_NAME`;
- `BRAND_PRIMARY_COLOR`;
- `BRAND_SECONDARY_COLOR`;
- `CLIENT_TIMEZONE`;
- `CLIENT_LOCALE`;
- `CLIENT_BRANDING_PATH`.

Run `deploy/control-plane/sync-client-branding.ps1` to write `branding.json` and update environment values.

## Domain routing

For public domains, generate a Caddy routing bundle from `deploy/control-plane`:

```powershell
.\generate-routing.ps1 -ClientSlug demo-client -Domain demo-client.example.org -FrontendTarget localhost:8088
```

Caddy handles HTTPS automatically after DNS points the domain to the routing host.

## Optional profiles

Redis is disabled by default. To start it with the client stack, set this in `.env.client`:

```text
COMPOSE_PROFILES=redis
```

To start background workers together with Redis:

```text
COMPOSE_PROFILES=redis,worker
```

Worker settings:

- `REDIS_URL`;
- `WORKER_QUEUE`;
- `WORKER_CONCURRENCY`.

Demo seeds are not applied automatically for production client stacks. To load seed data explicitly, run:

```powershell
.\seed.ps1
```

## Operations

```powershell
.\start.ps1
.\logs.ps1
.\stop.ps1
```

## Bootstrap admin

The `bootstrap-admin` job creates or updates the first administrator using:

- `BOOTSTRAP_ADMIN_EMAIL`;
- `BOOTSTRAP_ADMIN_PASSWORD`;
- `BOOTSTRAP_ADMIN_FIRST_NAME`;
- `BOOTSTRAP_ADMIN_LAST_NAME`;
- `BOOTSTRAP_ORGANIZATION_NAME`;
- `BOOTSTRAP_ORGANIZATION_SLUG`.

The user receives `super_admin` membership in the default organization and `system_admin` system role.

## Backup

```powershell
.\backup.ps1
```

The script creates a timestamped folder with:

- `database.sql`;
- `uploads.tar.gz`;
- `env.snapshot` with secrets redacted;
- `manifest.json`.

To register backup metadata in control plane:

```powershell
.\backup.ps1 -ControlPlaneApiKey "change-me-control-plane-key" -Register
```

Encrypted backup:

```powershell
.\backup.ps1 -Encrypt
.\restore.ps1 -BackupPath .\backups\20260517_120000 -EncryptionKey "secret-key"
```

## Export and import

Create a migration archive:

```powershell
.\export-client.ps1 -ControlPlaneApiKey "change-me-control-plane-key" -Register
```

Restore from an archive:

```powershell
.\import-client.ps1 -ArchivePath .\exports\demo-client_20260517_120000
```

## Restore

```powershell
.\restore.ps1 -BackupPath .\backups\20260517_120000
```

Use restore carefully: it replaces database state with the selected dump.
