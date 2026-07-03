# Optional FreeLLMAPI Private Image Mirror

This directory is now optional legacy/private-mirror infrastructure.

Upstream FreeLLMAPI publishes an official image:

```text
ghcr.io/tashfeenahmed/freellmapi:latest
```

Use that upstream image directly unless you specifically want to mirror/build images into your own Forgejo registry.

## When this repo still makes sense

Use the builder only if:

- Portainer should pull only from your Forgejo registry;
- you need internal image promotion or auditing;
- GHCR access is blocked/unwanted from your deployment network;
- you want to pin and mirror upstream commits under private tags.

Application/project repos must not build, install, upgrade, or administer FreeLLMAPI.

## Proven private-mirror architecture

```text
Forgejo repo: freellmapi-image-builder
  -> Forgejo Docker runner
  -> internal Forgejo registry push endpoint
  -> Forgejo container registry
  -> Portainer manually pulls public/LAN registry endpoint
  -> one shared FreeLLMAPI service for all projects
```

## Required repo variables

```text
DOCKER_HOST=tcp://172.23.0.2:2375
REGISTRY_PUSH_HOST=192.168.0.78:3205
REGISTRY_PULL_HOST=git.neomod.cc
IMAGE_NAMESPACE=neomod
IMAGE_NAME=freellmapi
KEEP_FREELLMAPI_VERSIONS=3
```

Adapt hostnames/IPs to your environment.

## Forgejo workflow gotchas

- Keep `REGISTRY_PUSH_HOST` runner-reachable. Use `REGISTRY_PULL_HOST` for Portainer/client pulls.
- Paste repo variables carefully. Trailing `\r` or `\n` can corrupt Docker tags or Forgejo API URLs.
- This template now strips trailing newlines defensively, but clean variable values are still preferred.
- When composing Docker tags in Forgejo shell steps, recompute critical refs from current vars inside each step instead of relying on fragile cross-step image-name handoff.

## Required repo secrets

```text
REGISTRY_USER=<Forgejo user or bot user>
REGISTRY_TOKEN=<token with package/container registry permission>
```

## Critical safety rules

The builder workflow must never run global Docker cleanup commands because your Forgejo Docker runner uses a shared remote Docker daemon through `DOCKER_HOST`.

Forbidden:

```bash
docker image prune
docker builder prune
docker system prune
```

Allowed:

```text
Remove only exact FreeLLMAPI tags created by the current workflow run.
```

## Native build dependencies

FreeLLMAPI depends on `better-sqlite3`. If no prebuilt binary exists for the current Node/runtime combo, it falls back to `node-gyp`. The generated builder-stage Dockerfile must install:

```text
python3
make
g++
```

Keep those tools in the builder stage only.

## Registry cleanup

The tested cleanup workflow:

- runs automatically;
- sanitizes Forgejo host/package variables before composing API URLs;
- uses only `GET /api/v1/packages/{owner}?type=container`;
- keeps latest `KEEP_FREELLMAPI_VERSIONS` `main-*` versions;
- deletes only older `main-*` and matching `sha-*` versions;
- never deletes `latest`, `sha256:*`, unrelated packages, Docker images, or Docker cache.
