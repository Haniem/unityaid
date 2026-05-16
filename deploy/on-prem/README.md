# UnityAid on-prem package

This folder contains tooling for preparing a customer-owned deployment bundle.

## Build package

```powershell
.\package-onprem.ps1 -AppVersion "2026.05"
```

The generated package contains:

- client Docker Compose stack;
- `.env.client.example`;
- migration and bootstrap jobs;
- backup/restore scripts;
- export/import scripts;
- documentation and package manifest.

## Install on customer server

1. Copy the generated package to the server.
2. Copy `client/.env.client.example` to `client/.env.client`.
3. Fill secrets, ports, domain, SMTP and bootstrap admin values.
4. Run `client/start.ps1`.
5. Configure DNS and TLS termination through Caddy, nginx, ingress or an existing customer gateway.

By default the on-prem package does not require control plane connectivity.
