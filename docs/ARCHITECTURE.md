# Architecture

This document describes the architecture of the project and the agentic tooling around it. For the detailed editor/agent endpoint split, see `docs/TOOLING_MODEL.md`.

## Tooling architecture

```text
Developer workstation
├─ VS Code
│  └─ human editor/workbench
├─ OpenCode
│  ├─ main configurable local agent
│  ├─ DeepSeek default cheap frontier route
│  ├─ FreeLLMAPI small/free route
│  ├─ optional OpenAI API backup route
│  └─ GitNexus MCP access
├─ GitNexus
│  └─ required repository memory/index layer
├─ Caveman Code
│  └─ separate low-token terminal/Codex-capable coding endpoint
├─ optional Codex in VS Code
│  └─ IDE-integrated Codex endpoint if preferred
├─ optional Caveman Skill
│  └─ brevity integration for selected agents/editors
└─ Forgejo
   └─ canonical Git remote
```


## Tooling startup sequence

```text
Verify remotes and current upstream sync
Deploy or verify FreeLLMAPI
Set up free models
VS Code
OpenCode
Load .env through start-opencode script
GitNexus install/index
GitNexus MCP verification
Caveman Code terminal endpoint
Paid/frontier OpenCode model routes
OpenCode planning
Caveman Code implementation
```

## Repository layout

This project keeps upstream application files at repository root to stay close to `redenfire/docker-image-watcher` and reduce divergence for future upstream contributions.

Primary application files and folders:

```text
main.go
docker.go
registry.go
web/
Dockerfile
docker-compose.yml
```

Agentic tooling, project docs, memory, deployment helpers, and editor-specific config stay in their dedicated template folders such as `docs/`, `memory/`, `scripts/`, `tools/`, `.cave/`, `.opencode/`, `.claude/`, `.forgejo/`, and `deploy/`.

## Project architecture

### System overview

`docker-image-watcher` is a small Go HTTP service with embedded web assets. It inspects running Docker containers through Docker Engine API, resolves remote image manifests from OCI-compatible registries, compares local and remote digests, and exposes a web UI plus API endpoints for status and update actions.

### Main components

| Component | Responsibility | Notes |
|---|---|---|
| `main.go` | HTTP server, routing, periodic refresh loop, auto-update persistence, image status state | Embeds `web/` assets into binary |
| `docker.go` | Docker Engine API access, container listing/inspection, image pulls, container recreation | Uses Unix socket at `/var/run/docker.sock` |
| `registry.go` | OCI/Docker registry manifest lookup and auth token handling | Compares remote digests against local images |
| `web/` | HTML/CSS/JS templates and frontend assets | Served from embedded filesystem |

### Data flow

1. User opens web UI.
2. Browser requests `/api/images`.
3. Server lists running containers through Docker Engine API.
4. Server resolves local image digest and remote registry digest for each image.
5. UI shows current state and outdated containers.
6. User triggers update or enables auto-update.
7. Server pulls image, recreates container, and refreshes status.

### Runtime/deployment model

- Service runs as Go binary or Docker container.
- Typical deployment mounts `/var/run/docker.sock` into container.
- Web UI is served by same Go process on configurable `PORT`.
- Auto-update preferences persist through `AUTO_FILE`, defaulting to `/data/auto-update.json`.

### External dependencies

- Docker Engine API
- OCI-compatible image registry
- Docker registry auth endpoints when registry requires bearer tokens

## Architectural decisions

Record durable decisions in `docs/DECISIONS.md`.
