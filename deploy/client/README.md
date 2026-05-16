# UnityAid client stack

This folder contains the first production-oriented client stack template for the managed single-tenant model.

## Quick start

1. Copy `.env.client.example` to `.env.client`.
2. Fill secrets, domains, ports and bootstrap admin values.
3. Start the stack:

```powershell
docker compose --env-file .env.client -f docker-compose.client.yml up -d --build
```

The stack creates:

- PostgreSQL database;
- migration job;
- bootstrap admin job;
- backend API;
- production frontend served by nginx;
- persistent uploads volume.

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
- `env.snapshot`.

## Restore

```powershell
.\restore.ps1 -BackupPath .\backups\20260517_120000
```

Use restore carefully: it replaces database state with the selected dump.
