# Troubleshooting

| Symptom | Likely cause | Solution |
|---|---|---|
| All images show `unknown` status | Docker socket not mounted, wrong socket path, or insufficient permissions | Check `/var/run/docker.sock` mount and `DOCKER_SOCK` value |
| Some images show `unknown` | Registry unreachable, registry auth required, or digest lookup failed | Check network connectivity and verify host can pull image |
| Pull fails / `pull error` | Docker Hub rate limit, private registry auth issue, or Docker daemon error | Set `DOCKER_REGISTRY_AUTH=username:password` for Docker Hub, or use JSON multi-registry auth when multiple registries need separate credentials. Run `docker pull` on host to verify auth, and wait for rate-limit reset if still anonymous |
| Persistent container error text is visible | Last pull attempt failed and error is being kept for operator visibility | Fix registry auth or connectivity, then retry the container update. Error clears after a later successful pull |
| Rate-limit banner is visible | Recent check or pull hit anonymous registry throttling | Configure `DOCKER_REGISTRY_AUTH`, then wait for the next successful check or pull to clear the banner |
| Update button disabled | Container is already `uptodate` or status is `unknown` | Only `outdated` containers can be updated from UI |
| Auto-update not triggering | Cooldown active, container not outdated, or toggle not persisted | Wait 5 minutes, confirm status is `outdated`, verify `/data` persistence |
| Progress bar stuck at `0%` | Large image still downloading or daemon not yet reporting layer totals | Wait longer and inspect Docker daemon activity |
| Container not recreated after pull | Create/start failed due to port conflict, bad config, or incompatible new image | Check application logs and Docker daemon errors |

## Quick Checks

### Verify health endpoint

```bash
curl -i http://localhost:8099/health
```

### Verify API list

```bash
curl http://localhost:8099/api/images
```

### Inspect container logs

```bash
docker logs image-watch
```

### Verify Docker socket mount

```bash
docker inspect image-watch --format '{{json .Mounts}}'
```

### Verify host can pull target image

```bash
docker pull nginx:latest
```

### Retry with authenticated pull config

```bash
DOCKER_REGISTRY_AUTH=username:password docker compose up -d
DOCKER_REGISTRY_AUTH='{"ghcr.io":"user:token","https://index.docker.io/v1/":"user:token"}' docker compose up -d
```

## Common Causes of `unknown`

`unknown` status happens when local digest or remote digest cannot be resolved. Common reasons:

- image started from raw `sha256:` reference
- registry blocks anonymous manifest requests
- registry auth challenge/token exchange fails
- Docker image metadata has no `RepoDigests`
