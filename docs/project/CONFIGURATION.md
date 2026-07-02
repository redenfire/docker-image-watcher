# Configuration

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port for web UI and API |
| `DOCKER_SOCK` | `/var/run/docker.sock` | Unix socket path used for Docker Engine API |
| `AUTO_FILE` | `/data/auto-update.json` | JSON file storing per-container auto-update state |

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

### `AUTO_FILE`

- Read in `main.go`
- Defaults to `/data/auto-update.json`
- Stores JSON map like:

```json
{
  "3d4c5b6a7f8e": true,
  "9f8e7d6c5b4a": false
}
```

- Persist `/data` with volume mount to keep toggle state across restarts

## Docker Compose Notes

Current `docker-compose.yml`:

- publishes `8099:8080`
- mounts Docker socket at `/var/run/docker.sock`
- mounts named volume `image-watch-data:/data`
- sets `PORT`, `DOCKER_SOCK`, and `AUTO_FILE`

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
      - image-watch-data:/data
    environment:
      - PORT=8080
      - DOCKER_SOCK=/var/run/docker.sock
      - AUTO_FILE=/data/auto-update.json
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

### Data volume

```text
image-watch-data:/data
```

Recommended so auto-update preferences survive container recreation and host restarts.

## Registry Access Tips

- Public images typically need no extra config.
- Private registries depend on Docker daemon auth state.
- If host can already `docker pull` private image successfully, Image Watch usually inherits that access because pull happens through Docker Engine.
