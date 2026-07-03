# Project Constraints

This file records constraints for this specific project.

It is not for workstation setup, agent installation, or template maintenance notes.

## Business constraints

- Project must monitor Docker containers for outdated images.
- UI and update flow should stay minimal and operator-focused.
- Changes should keep path open for sending improvements upstream.

## Technical constraints

- Go-based application.
- Runtime depends on Docker Engine API through mounted Docker socket.
- Multi-arch Docker build should remain supported.
- Current upstream application layout keeps main Go files at repository root.

## Security constraints

- `/var/run/docker.sock` mount is privileged and increases host-control risk.
- Treat local agent config and API credentials as local-only material.
- Avoid committing generated indexes or secret-bearing local config.

## Operational constraints

- App is expected to run as a containerized service.
- Auto-update state must persist across restarts.
- Docker registry access depends on reachable OCI endpoints and auth flows.

## Compatibility constraints

- Must work with OCI-compatible registries.
- Must work with Docker Engine environments that expose the Unix socket.

## Budget / token constraints

- FreeLLMAPI should be preferred for low-risk work when adequate.
- OpenAI API backup usage must be explicit if personal API credit is involved.
- Keep agent/tooling overhead low during routine maintenance.

## Non-goals

- Turning project into full container orchestration platform.
- Introducing unrelated template refactors during initialization.
