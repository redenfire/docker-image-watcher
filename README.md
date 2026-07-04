# 🐳 Docker Image Watcher

Minimal Docker image update monitor with web UI. Checks running containers against their registry, shows outdated images, and can auto-update.

## Quick Start

```bash
docker run -d \
  --name image-watch \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v image-watch-data:/data \
  -p 8099:8080 \
  ghcr.io/redenfire/docker-image-watcher:latest
```

Open [http://localhost:8099](http://localhost:8099).

Or use Docker Compose:

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
      # - AUTH_USER=admin       # uncomment to enable auth
      # - AUTH_PASS=changeme

volumes:
  image-watch-data:
```

## Features

- Lists running containers grouped by image, with per-container status
- Supports OCI-compatible registries including Docker Hub, GHCR, Quay, and self-hosted registries
- One-click update per container or "Update all" per image group, with live progress bars
- Per-container auto-update toggle persisted to disk
- Background check every 10 minutes with 5-minute auto-update cooldown
- i18n with EN/IT language toggle
- Optional auth via AUTH_USER/AUTH_PASS with login page and HMAC-signed session cookies
- Excludes own container from listing and blocks self-update
- Multi-arch container build support for `linux/amd64`, `linux/arm64`, `linux/arm/v7`

## How It Works

Three Go modules work together:

| Module | Responsibility |
|---|---|
| `main.go` | HTTP server, routes, periodic checks, auto-update state |
| `docker.go` | Docker Engine API client, image pulls, container recreation |
| `registry.go` | Registry manifest lookup and token-based auth |

Check cycle:

```text
main.go:checkAll()
-> docker.go:listContainers()
-> for each image group:
   -> registry.go:getRemoteDigest()
   -> for each container:
      -> docker.go:getImageDigest()  // fallback: getLocalDigest()
      -> compare digests
-> refresh ImageGroup list
-> auto-update outdated containers when enabled
```

Update cycle:

```text
main.go:updateContainer()
-> docker.go:pullImageStream()
-> docker.go:recreateContainer()
-> main.go:checkAll()
```

## API

| Endpoint | Purpose |
|---|---|
| `GET /api/images` | List tracked running containers grouped by image |
| `POST /api/images/{id}/update` | Trigger async pull + recreate for one container |
| `POST /api/images/{id}/auto-update` | Toggle per-container auto-update |
| `GET /api/images/{id}/progress` | Read pull/update progress |
| `POST /api/groups/update` | Trigger async pull + recreate for all containers of an image |
| `POST /api/login` | Authenticate and receive session cookie |
| `POST /api/logout` | Invalidate session cookie |
| `GET /api/auth/status` | Return whether auth is enabled |
| `GET /health` | Health check |

Full reference: [`docs/project/API.md`](docs/project/API.md)

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DOCKER_SOCK` | `/var/run/docker.sock` | Docker Unix socket path |
| `AUTO_FILE` | `/data/auto-update.json` | Persisted auto-update state file |
| `AUTH_USER` | — | Enable HTTP auth (required together with AUTH_PASS) |
| `AUTH_PASS` | — | Password for HTTP auth |
| `CHECK_INTERVAL` | `10m` | Background check interval |
| `CHECK_CONCURRENCY` | `5` | Max concurrent registry requests during check |
| `AUTO_COOLDOWN` | `5m` | Cooldown between auto-updates for same container |

Details: [`docs/project/CONFIGURATION.md`](docs/project/CONFIGURATION.md)

## Auto-Update

If auto-update is enabled for a container, Image Watch checks every 10 minutes and automatically pulls plus recreates the container when a newer image digest is available. A 5-minute cooldown prevents repeated rapid updates.

Details: [`docs/project/AUTO-UPDATE.md`](docs/project/AUTO-UPDATE.md)

## Build

```bash
# Local Go build (requires Go 1.22+)
go build -o image-watch .

# Multi-arch Docker build
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t ghcr.io/redenfire/docker-image-watcher:latest \
  --push .
```

Details: [`docs/project/BUILD.md`](docs/project/BUILD.md)

## Security

> [!WARNING]
> This container requires access to `/var/run/docker.sock`. That is effectively root-equivalent access to Docker host.

When exposing Image Watch on a network, set `AUTH_USER` and `AUTH_PASS` to enable authenticated access with a login page and HMAC-signed session cookies. Login brute-force is rate-limited (5 failures/min → 30s block).

Mitigations and deployment guidance: [`docs/project/SECURITY.md`](docs/project/SECURITY.md)

## Health Checks

The application exposes a `GET /health` endpoint that returns HTTP 200. Since the image is based on `scratch` for minimal size, Docker `HEALTHCHECK` is not built into the image.

**Portainer stacks:** Configure an external HTTP health check pointing to `http://<container-name>:8080/health` in the Portainer UI, or use Portainer's HTTP ping capability to monitor the container from outside.

**Other orchestrators:** Any standard HTTP health probe targeting `http://<container-ip>:8080/health` works.

## Troubleshooting

- Images showing `unknown` status: check Docker socket access and registry connectivity
- Pull failures: check Docker Hub rate limits or private registry auth state
- Auto-update not triggering: confirm container is `outdated` and not inside cooldown window

More: [`docs/project/TROUBLESHOOTING.md`](docs/project/TROUBLESHOOTING.md)

## Contributing

1. Fork upstream repository or clone from Forgejo instance.
2. Create feature branch: `git checkout -b feature/my-change`
3. Make change.
4. Open pull request against upstream repository when appropriate.

## License

MIT — see [`LICENSE`](LICENSE).
