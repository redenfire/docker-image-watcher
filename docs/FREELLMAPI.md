# FreeLLMAPI

FreeLLMAPI is shared infrastructure, not a per-project dependency.

Use it as the low-risk/free route for:

- documentation cleanup;
- summaries;
- formatting;
- boilerplate;
- low-risk review;
- small/background model tasks in OpenCode.

Do not use it as the only route for high-risk architecture, security-sensitive changes, or hard production implementation unless the user explicitly accepts the risk.

## Deployment

Use the upstream Docker image path documented in `deploy/freellmapi/`.

Important deployment rules:

- `ENCRYPTION_KEY` must be a 64-character hexadecimal string.
- Generate it once during first deployment.
- Keep it permanently.
- Do not regenerate after database creation.

Generate:

```bash
openssl rand -hex 32
```

PowerShell:

```powershell
-join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Minimum 0 -Maximum 256) })
```

## Model catalog verification

The FreeLLMAPI auto-router model id is:

```text
auto
```

Do not use the older template placeholder `free-auto`. If OpenCode reports:

```text
Model 'free-auto' is not in the catalog. Use 'auto' ...
```

then update `opencode.json` so the model list contains `auto` and `small_model` is `freellmapi/auto`.

Verify the live catalog with:

```powershell
curl.exe "$env:FREELLMAPI_BASE_URL/models" -H "Authorization: Bearer $env:FREELLMAPI_API_KEY"
```

Linux/macOS:

```bash
curl "$FREELLMAPI_BASE_URL/models" \
  -H "Authorization: Bearer $FREELLMAPI_API_KEY"
```

## OpenCode route

`opencode.json` uses:

```json
"small_model": "freellmapi/auto"
```

FreeLLMAPI is also available through `/models` if env values are loaded.

Verify with:

```powershell
opencode debug config
```

`baseURL` and `apiKey` must not be empty.
