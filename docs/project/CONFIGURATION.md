# Configuration

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port for web UI and API |
| `DOCKER_SOCK` | `/var/run/docker.sock` | Unix socket path used for Docker Engine API |
| `CHECK_INTERVAL` | `10m` | Interval for background digest check loop |
| `CHECK_CONCURRENCY` | `5` | Max concurrent registry requests during check |
| `AUTO_COOLDOWN` | `5m` | Cooldown window between auto-updates for same container |

### `AUTH_USER` / `AUTH_PASS`

- Read in `main.go`
- When both set, enables HMAC-signed session auth with login page
- Unauthenticated requests to API return 401
- Login brute-force rate-limited: 5 failures/min → 30s block

### `CHECK_INTERVAL`

- Read in `main.go` via `getEnvDuration`
- Defaults to `10m` (10 minutes)
- Controls how often the background goroutine checks all running containers against registry

### `CHECK_CONCURRENCY`

- Read in `main.go` via `os.Getenv`
- Defaults to `5`
- Caps parallel registry manifest requests during a check cycle

### `AUTO_COOLDOWN`

- Read in `main.go` via `getEnvDuration`
- Defaults to `5m` (5 minutes)
- Prevents repeated auto-updates for the same container within the window

## Variable Details

### `PORT`

- Read in `main.go`
- Defaults to `8080`
- Exposed as container port `8080` in `Dockerfile`
- Common deployment maps host port `8099` to container port `8080`

### `DOCKER_SOCK`

- Read during process init in `docker.go`
- Defaults to `/var/run/docker.sock`
- Useful for custom Unix socket paths or rootless Docker setups where socket lives elsewhere
- Compose file passes `DOCKER_SOCK=/var/run/docker.sock` explicitly

### `AUTO_COOLDOWN`

- Read in `main.go` via `getEnvDuration`
- Defaults to `5m` (5 minutes)
- Prevents repeated auto-updates for the same container within the window

## Docker Label for Auto-Update

Auto-update is enabled via the Docker label `image-watch.auto-update` on the container:

```yaml
services:
  my-service:
    labels:
      - "image-watch.auto-update=true"
```

This label survives recreate because `recreateContainer()` copies labels from the old container. See [`AUTO-UPDATE.md`](AUTO-UPDATE.md) for details.

## Docker Compose Notes

Current `docker-compose.yml`:

- publishes `8099:8080`
- mounts Docker socket at `/var/run/docker.sock`
- sets `PORT` and `DOCKER_SOCK`

Example:

```yaml
services:
  image-watch:
    image: ghcr.io/redenfire/docker-image-watcher:latest
    container_name: image-watch
    restart: unless-stopped
    ports:
      - "8099:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - PORT=8080
      - DOCKER_SOCK=/var/run/docker.sock
```

## Volume Mounts

### Docker socket mount

```text
/var/run/docker.sock:/var/run/docker.sock
```

Required so Image Watch can:

- list running containers
- inspect container config
- pull images
- stop/remove/create/start containers

The previous `image-watch-data:/data` data volume is no longer needed since auto-update state is read from Docker labels, not from a file.

## Registry Access Tips

- Public images typically need no extra config.
- Private registries depend on Docker daemon auth state.
- If host can already `docker pull` private image successfully, Image Watch usually inherits that access because pull happens through Docker Engine.
