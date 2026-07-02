# Troubleshooting

| Symptom | Likely cause | Solution |
|---|---|---|
| All images show `unknown` status | Docker socket not mounted, wrong socket path, or insufficient permissions | Check `/var/run/docker.sock` mount and `DOCKER_SOCK` value |
| Some images show `unknown` | Registry unreachable, registry auth required, or digest lookup failed | Check network connectivity and verify host can pull image |
| Pull fails / `pull error` | Docker Hub rate limit, private registry auth issue, or Docker daemon error | Run `docker pull` on host first to verify auth and wait for rate-limit reset if needed |
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

## Common Causes of `unknown`

`unknown` status happens when local digest or remote digest cannot be resolved. Common reasons:

- image started from raw `sha256:` reference
- registry blocks anonymous manifest requests
- registry auth challenge/token exchange fails
- Docker image metadata has no `RepoDigests`
