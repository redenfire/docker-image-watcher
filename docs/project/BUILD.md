# Build and Run

## Prerequisites

- Go `1.22+`
- Docker for container builds
- Docker Buildx for multi-arch builds

## Local Go Build

```bash
go build -o image-watch .
```

## Run Locally

Linux example:

```bash
DOCKER_SOCK=/var/run/docker.sock AUTO_FILE=./auto-update.json PORT=8080 ./image-watch
```

Requirements:

- process must reach valid Docker Unix socket
- local user must have permission to use that socket
- registry network access must be available

## Docker Build

Single-platform image:

```bash
docker build -t image-watch .
```

## Multi-Arch Docker Build

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t ghcr.io/redenfire/docker-image-watcher:latest \
  --push .
```

## Cross-Compilation

Example Linux ARM64 binary:

```bash
GOOS=linux GOARCH=arm64 go build -o image-watch-linux-arm64 .
```

## Dockerfile Notes

Current `Dockerfile` uses multi-stage build:

1. `golang:1.22-alpine` build stage compiles static binary with `CGO_ENABLED=0`
2. build stage installs `ca-certificates`
3. final stage uses `scratch`
4. final image contains only binary and cert bundle

This keeps runtime image small while still allowing HTTPS registry calls.

## Embedded Web Assets

Frontend is embedded into binary through Go directive in `main.go`:

```go
//go:embed web
var webFS embed.FS
```

That means `web/index.html` ships inside built binary and does not need separate runtime mount.
