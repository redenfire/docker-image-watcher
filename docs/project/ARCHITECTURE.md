# Project Architecture

## Overview

Image Watch is a Go 1.22 single-binary web service. It uses Go standard library packages for HTTP serving, JSON handling, embedded static assets, registry requests, and Docker Engine communication over Unix socket.

Runtime dependencies:

- Docker Engine socket
- Reachable OCI-compatible registries

## Component Diagram

```text
+-------------------+       +-------------------+       +------------------+
|   Web Browser     |       |    Image Watch    |       |  Docker Engine   |
|   (index.html)    | HTTP  |    (Go binary)    | Unix  |    (daemon)      |
|                   +------>+                   +------>+                  |
| - Container list  |       | - main.go         | sock  | - list/inspect   |
| - Update button   |       | - docker.go       |       | - pull stream    |
| - Progress bar    |       | - registry.go     |       | - create/start   |
| - Auto toggle     |       |                   |       | - stop/remove    |
+-------------------+       +--------+----------+       +------------------+
                                      |
                                      | HTTPS
                                      v
                               +------------------+
                               |   OCI Registry   |
                               | Docker Hub/GHCR  |
                               | Quay/self-hosted |
                               +------------------+
```

## Data Structures

| Structure | Defined in | Purpose |
|---|---|---|
| `ImageStatus` | `main.go` | API/UI state for one running container |
| `dockerContainer` | `docker.go` | Container list entry from Docker Engine |
| `dockerInspect` | `docker.go` | Partial inspect payload used for recreation |
| `PullProgress` | `docker.go` | Aggregated pull progress reported to UI |

## Main Components

### `main.go`

- Embeds `web/index.html` into binary via `//go:embed`
- Creates HTTP routes
- Starts periodic `checkAll()` loop every 10 minutes
- Maintains shared state for images, auto-update settings, cooldowns, and progress
- Performs manual and automatic update orchestration

### `docker.go`

- Builds Docker API client over Unix socket
- Lists running containers
- Reads local image digests from Docker image metadata
- Streams pull progress and aggregates layer totals
- Recreates containers by inspect -> stop -> remove -> create -> start
- Honors `DOCKER_SOCK` environment variable at process startup

### `registry.go`

- Parses image references into registry/repository/tag
- Fetches remote image manifests
- Handles bearer-token auth challenge flow
- Returns `Docker-Content-Digest` for comparison with local digest

### `web/index.html`

- Polls `/api/images` every 10 seconds
- Shows container table, digest comparison, status badges, progress UI, and auto-update toggles
- Calls update and auto-update API routes directly from browser JavaScript

## Key Call Chains

### Check cycle

```text
main.go:checkAll()
-> docker.go:listContainers()
-> docker.go:getLocalDigest()
-> registry.go:getRemoteDigest()
-> compare digests
-> store []ImageStatus
-> if auto_update && outdated && cooldown expired:
   -> main.go:updateContainer()
```

### Update cycle

```text
main.go:updateContainer()
-> docker.go:pullImageStream()
-> docker.go:recreateContainer()
   -> docker.go:inspectContainer()
   -> Docker stop
   -> Docker remove
   -> Docker create
   -> Docker start
-> main.go:checkAll()
```

### Auto-update persistence

```text
main.go:toggleAuto()
-> update app.images
-> update app.autoSaved
-> saveAuto()
-> write AUTO_FILE JSON map
```

## Route-to-Handler Mapping

| Route | Method | Handler | Behavior |
|---|---|---|---|
| `/` | `GET` | `http.FileServer(http.FS(sub))` | Serves embedded `web/` assets |
| `/api/images` | `GET` | `app.handleImages` | Returns JSON array of `ImageStatus` |
| `/api/images/{id}/update` | `POST` | `app.handleImageAction` | Starts async update goroutine |
| `/api/images/{id}/auto-update` | `POST` | `app.handleImageAction` | Toggles persisted auto-update flag |
| `/api/images/{id}/progress` | `GET` | `app.handleImageAction` | Returns `PullProgress` for in-flight update |
| `/health` | `GET` | inline handler | Returns HTTP 200 with empty body |

## Concurrency Model

- `sync.RWMutex` protects `images`, `autoSaved`, and `cooldowns`
- `sync.Map` stores per-container progress snapshots
- Manual update runs in goroutine from route handler
- Periodic checks run in background ticker goroutine
- Auto-update cooldown prevents rapid repeated updates for same container
