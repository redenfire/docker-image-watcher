# Image Watch

Minimal Docker image update monitor with a web UI. Checks running containers against their registry, shows outdated images, and can auto-update.

## Usage

```bash
docker run -d \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 8099:8080 \
  ghcr.io/<your-user>/image-watch:latest
```

Open `http://localhost:8099`.

## Features

- Lists all running containers with local vs remote digest comparison
- Supports any OCI registry (Docker Hub, GHCR, quay.io, self-hosted)
- One-click **Update** with live progress bar
- Per-container **Auto-update** toggle (persisted on disk)
- Auto-update checks every 10 minutes with 5 min cooldown
- Multi-arch: `linux/amd64`, `linux/arm64`, `linux/arm/v7`

## API

| Endpoint | Method | Description |
|---|---|---|
| `/api/images` | GET | List monitored images with status |
| `/api/images/{id}/update` | POST | Pull + recreate container |
| `/api/images/{id}/auto-update` | POST | Toggle auto-update |
| `/api/images/{id}/progress` | GET | Pull progress (poll while updating) |
| `/health` | GET | Health check |

## Build

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t ghcr.io/<your-user>/image-watch:latest --push .
```

## Env

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Web UI port |
| `DOCKER_SOCK` | `/var/run/docker.sock` | Docker socket path |
| `AUTO_FILE` | `/data/auto-update.json` | Auto-update config persistence |