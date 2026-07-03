# Project Brief

## Name

docker-image-watcher (Image Watch)

## Purpose

Minimal Docker image update monitor with web UI. Checks running containers against their registry, shows outdated images, and can auto-update container refreshes.

## Users

Docker operators managing self-hosted containers.

## Success criteria

- Container list loads from live Docker Engine data.
- Update flow recreates a container on newer remote image digest.
- Auto-update toggles persist and work reliably.
- Docker image and compose flow support multi-arch builds.

## Constraints

- Private/self-hosted Git source: Forgejo.
- Prefer local-first tooling where possible.
- Avoid committing generated local indexes.
- Runtime requires Docker socket access.
- Keep codebase close to upstream project to ease future upstream contributions.

## Current status

Project code imported from upstream `redenfire/docker-image-watcher`, agentic template restored on top, TASK-001 initialization in progress.
