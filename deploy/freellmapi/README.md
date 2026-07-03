# FreeLLMAPI Shared Service

FreeLLMAPI is treated as shared infrastructure, not as a per-project dependency.

## Current standard path

Upstream FreeLLMAPI now publishes a production container image:

```text
ghcr.io/tashfeenahmed/freellmapi:latest
```

Use that image directly in Portainer unless you have a strong reason to mirror images into your private Forgejo registry.

## Correct model

```text
One permanent Portainer stack
Many project repos connect to the shared service
```

Application projects should only configure:

```env
FREELLMAPI_BASE_URL=http://freellmapi.example.lan:3001/v1
FREELLMAPI_API_KEY=<shared FreeLLMAPI proxy key>
```

They must not install, build, upgrade, or administer FreeLLMAPI.

## Included directories

```text
portainer/
  Stack example and deployment notes for the shared FreeLLMAPI service.

builder-repo/
  Optional legacy/private-mirror template. Not required for normal FreeLLMAPI use now that upstream publishes GHCR images.
```

## ENCRYPTION_KEY

Generate one permanent 64-character hexadecimal key on first creation and keep it across upgrades.

Linux/macOS:

```bash
openssl rand -hex 32
```

Windows PowerShell:

```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Minimum 0 -Maximum 256) })
```

Changing the key after provider credentials are stored may make those credentials unusable.

## Optional private mirror

The old Forgejo image-builder workflow remains in `builder-repo/` as a private mirror pattern. Use it only if:

- you want all runtime images stored in Forgejo registry;
- your Portainer host cannot or should not pull from GHCR;
- you want internal image promotion policy.

Forgejo mirror note:

- keep runner push host and client pull host separate when needed;
- sanitize repo variables before composing Docker tags or Forgejo API URLs;
- prefer recomputing critical image refs inside each shell step instead of depending on fragile image-name handoff across steps.

Otherwise, skip it.
