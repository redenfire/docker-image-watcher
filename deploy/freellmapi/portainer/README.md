# Portainer FreeLLMAPI Stack

This stack runs the shared FreeLLMAPI service.

## Preferred image

Use the upstream official image:

```text
ghcr.io/tashfeenahmed/freellmapi:latest
```

If upstream publishes version tags, prefer a pinned version for stable production use. Use `latest` when you accept default-branch updates.

## Persistent data

SQLite data lives at:

```text
/app/server/data
```

The compose file maps this to the named volume:

```text
freellmapi-data
```

Keep the volume when upgrading.

## ENCRYPTION_KEY

`ENCRYPTION_KEY` must be a 64-character hexadecimal string. Generate it once on first deployment and keep it permanently.

Linux/macOS:

```bash
openssl rand -hex 32
```

Windows PowerShell:

```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Minimum 0 -Maximum 256) })
```

Set it in Portainer using `stack.env.example` or Portainer environment variables:

```env
FREELLMAPI_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

Do not regenerate it after provider keys have been added.

## Update policy

Recommended manual update flow:

```text
1. Check upstream FreeLLMAPI README/releases/changelog.
2. Update the image tag if needed.
3. Redeploy the Portainer stack.
4. Test /v1/models and /v1/chat/completions through project health checks.
```

## LAN exposure

Expose FreeLLMAPI only on trusted networks or behind a protected reverse proxy. FreeLLMAPI is single-user by design.

## Example health checks

From a project repo configured with `.env`:

Linux/macOS:

```bash
./scripts/check-freellmapi.sh
```

Windows PowerShell:

```powershell
.\scripts\check-freellmapi.ps1
```

## Optional private mirror

If you use `deploy/freellmapi/builder-repo/`, use the mirrored Forgejo image in this compose file instead of GHCR, for example:

```text
git.neomod.cc/neomod/freellmapi:main-<sha>
```
