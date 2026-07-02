# Security

## Docker Socket Access

Image Watch requires access to Docker Engine API. In container deployments that normally means:

```text
-v /var/run/docker.sock:/var/run/docker.sock
```

That access is effectively root-equivalent on Docker host because application can:

- inspect running containers
- pull arbitrary images
- stop/remove/recreate containers
- create new containers with host-mounted resources

Treat deployment as privileged and only run it in trusted environments.

## Network Exposure

By default service listens on all interfaces inside container. If only local host access is needed, bind published port to loopback:

```text
127.0.0.1:8099:8080
```

For multi-user environments, prefer reverse proxy with authentication in front of Image Watch.

## Hardening Ideas

Where compatible with deployment model:

- restrict access to trusted hosts and internal networks
- run behind nginx, Caddy, or Traefik with auth
- use read-only root filesystem
- provide writable data path separately for `AUTO_FILE`
- drop unnecessary Linux capabilities at container runtime

Example approach:

- `--read-only`
- writable volume or tmpfs for `/data`
- loopback-only bind or reverse-proxy exposure

## Registry Credentials

Image Watch does not manage registry credentials directly. Pulls are executed through Docker Engine, so private registry access usually depends on daemon/host auth state already established with Docker.

## Operational Guidance

- avoid exposing service directly to public internet
- assume anyone with UI/API access can trigger container recreation
- review container set on host before enabling auto-update broadly
