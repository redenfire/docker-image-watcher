# API Reference

Base path: `/`

## Data Types

### `ImageGroup`

Returned by `GET /api/images`.

```json
{
  "image": "nginx:latest",
  "remote_digest": "sha256:2222222222",
  "status": "outdated",
  "containers": [
    {
      "container_id": "3d4c5b6a7f8e",
      "container_name": "nginx",
      "local_digest": "sha256:1111111111",
      "status": "outdated",
      "auto_update": true
    }
  ]
}
```

Fields:

- `image`: normalized image reference, with `:latest` added when missing
- `remote_digest`: shortened remote digest or `unknown`
- `status`: `uptodate`, `outdated`, `partial`, or `unknown`
- `containers`: array of `ContainerItem`

### `ContainerItem`

Fields:

- `container_id`: first 12 chars of Docker container ID
- `container_name`: Docker container name without leading `/`
- `local_digest`: shortened local digest or `unknown`
- `status`: `uptodate`, `outdated`, or `unknown`
- `auto_update`: persisted toggle state for this container

### `PullProgress`

Returned by `GET /api/images/{id}/progress`.

```json
{
  "layer": "a3ed95caeb02",
  "current": 1048576,
  "total": 8388608,
  "percent": 12,
  "status": "Downloading"
}
```

`status` can also be values such as `connecting`, `recreating`, or `error: ...`.

## Endpoints

### `GET /api/images`

Return all running containers grouped by image.

Success response:

```json
[
  {
    "image": "nginx:latest",
    "remote_digest": "sha256:2222222222",
    "status": "outdated",
    "containers": [
      {
        "container_id": "3d4c5b6a7f8e",
        "container_name": "nginx",
        "local_digest": "sha256:1111111111",
        "status": "outdated",
        "auto_update": false
      }
    ]
  }
]
```

`curl` example:

```bash
curl http://localhost:8099/api/images
```

### `POST /api/images/{id}/update`

Trigger async pull and recreate for given container.

Success response:

```json
{"status":"updating"}
```

`curl` example:

```bash
curl -X POST http://localhost:8099/api/images/3d4c5b6a7f8e/update
```

### `POST /api/images/{id}/auto-update`

Toggle auto-update for given container.

Success response:

```json
{"auto_update":true}
```

`curl` example:

```bash
curl -X POST http://localhost:8099/api/images/3d4c5b6a7f8e/auto-update
```

### `GET /api/images/{id}/progress`

Return current progress snapshot for an in-flight update.

Success response:

```json
{
  "layer": "a3ed95caeb02",
  "current": 1048576,
  "total": 8388608,
  "percent": 12,
  "status": "Downloading"
}
```

When no update is active yet, endpoint returns `404` with `no progress` body.

`curl` example:

```bash
curl http://localhost:8099/api/images/3d4c5b6a7f8e/progress
```

### `GET /api/ratelimit`

Return whether the backend has detected Docker pull rate limiting during recent checks or pull attempts.

Success response:

```json
{"rate_limited": true}
```

`curl` example:

```bash
curl http://localhost:8099/api/ratelimit
```

### `POST /api/groups/update`

Trigger async pull and recreate for all containers of an image group.

Request body:

```json
{"image":"nginx:latest"}
```

Success response:

```json
{"status":"updating"}
```

`curl` example:

```bash
curl -X POST http://localhost:8099/api/groups/update \
  -H 'Content-Type: application/json' \
  -d '{"image":"nginx:latest"}'
```

### `POST /api/login`

Authenticate with credentials and receive a session cookie.

Request body:

```json
{"user":"admin","pass":"changeme"}
```

Success response:

- HTTP `200` with `Set-Cookie` header
- Body: `{"status":"ok"}`

### `POST /api/logout`

Invalidate session cookie.

Success response:

- HTTP `200` with expired `Set-Cookie`

### `GET /api/auth/status`

Return whether auth is enabled.

Success response:

```json
{"enabled": true}
```

### `GET /health`

Health probe.

Success response:

- HTTP `200`
- Empty body

`curl` example:

```bash
curl -i http://localhost:8099/health
```

## Error Responses

Common errors emitted by handlers:

| Status | Body | When |
|---|---|---|
| `400` | `bad path: /api/images/{id}/{action}\n` | Path after `/api/images/` does not contain exactly `{id}/{action}` |
| `400` | `unknown action\n` | Action is not `update`, `auto-update`, or `progress` |
| `404` | `container not found\n` | Container ID is not present in current `app.images` snapshot |
| `404` | `no progress\n` | No active progress snapshot for given container |
| `405` | `method not allowed\n` | Wrong HTTP method used for endpoint |

## Notes

- `GET /api/images` only includes running containers because Docker API call uses `all=false`.
- `GET /api/ratelimit` remains `true` until a later check cycle completes without detecting rate limiting.
- Containers started from image IDs like `sha256:...` are skipped.
- Digests in API are shortened to first 17 characters for display.
