# API Reference

Base path: `/`

## Data Types

### `ImageStatus`

Returned by `GET /api/images`.

```json
{
  "container_id": "3d4c5b6a7f8e",
  "container_name": "nginx",
  "image": "nginx:latest",
  "local_digest": "sha256:1111111111",
  "remote_digest": "sha256:2222222222",
  "status": "outdated",
  "auto_update": true
}
```

Fields:

- `container_id`: first 12 chars of Docker container ID
- `container_name`: Docker container name without leading `/`
- `image`: normalized image reference, with `:latest` added when missing
- `local_digest`: shortened local digest or `unknown`
- `remote_digest`: shortened remote digest or `unknown`
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

Return all running containers currently tracked by Image Watch.

Success response:

```json
[
  {
    "container_id": "3d4c5b6a7f8e",
    "container_name": "nginx",
    "image": "nginx:latest",
    "local_digest": "sha256:1111111111",
    "remote_digest": "sha256:2222222222",
    "status": "outdated",
    "auto_update": false
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
| `400` | `bad path: /api/images/{id}/{action}` | Path after `/api/images/` does not contain exactly `{id}/{action}` |
| `400` | `unknown action` | Action is not `update`, `auto-update`, or `progress` |
| `404` | `container not found` | Container ID is not present in current `app.images` snapshot |
| `404` | `no progress` | No active progress snapshot for given container |
| `405` | `method not allowed` | Wrong HTTP method used for endpoint |

## Notes

- `GET /api/images` only includes running containers because Docker API call uses `all=false`.
- Containers started from image IDs like `sha256:...` are skipped.
- Digests in API are shortened to first 17 characters for display.
