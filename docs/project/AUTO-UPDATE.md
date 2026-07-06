# Auto-Update

## How It Is Enabled

Auto-update is controlled via the Docker label `image-watch.auto-update` on the container.

To enable, add the label to your container or service:

```yaml
services:
  my-service:
    labels:
      - "image-watch.auto-update=true"
```

The label is **declarative**: it survives every recreate (manual or automatic) because `recreateContainer()` copies existing labels from the old container to the new one during `inspect -> stop -> remove -> create -> start`.

The web UI displays the current auto-update state (read from the label) but does not allow toggling — the label must be set in the container definition (e.g., `docker-compose.yml`).

The `POST /api/images/{id}/auto-update` endpoint still returns the current state, but does not modify it.

## Why Labels Instead of a File

| Approach | Survives recreate? | Declarative? |
|---|---|---|
| `auto-update.json` (previous) | No — new container ID = lost state | No |
| Docker label `image-watch.auto-update` (current) | Yes — labels are copied on recreate | Yes |

This makes auto-update a true lifecycle feature rather than a one-shot toggle.

## Check Interval

Background refresh loop in `main.go`:

- runs one initial `checkAll()` on startup
- then runs periodically, default 10 minutes
- configurable via `CHECK_INTERVAL` env var

## Cooldown

Automatic updates use per-container cooldown:

- default cooldown window: 5 minutes
- configurable via `AUTO_COOLDOWN` env var
- stored in `App.cooldowns`
- checked inside `checkAll()` before triggering auto-update

This avoids repeated rapid recreate cycles when a container stays reported as outdated.

## Trigger Conditions

Auto-update only triggers when all conditions are true:

- container appears in current running container list
- container has label `image-watch.auto-update=true`
- computed status is `outdated`
- no active cooldown (default 5 min, configurable via `AUTO_COOLDOWN`) blocks that container

## Concurrency and State Safety

- `sync.RWMutex` protects `images` and `cooldowns`
- `sync.Map` stores pull progress snapshots
- route-triggered manual updates run in goroutines
- cooldown tracking reduces risk of concurrent repeated auto-updates for same container

## Manual Update Behavior

Manual updates via `POST /api/images/{id}/update`:

- do not check cooldown before starting
- immediately launch `updateContainer()` in goroutine
- expose progress through `/api/images/{id}/progress`

## Update Sequence

```text
auto or manual trigger
-> updateContainer()
-> pullImageStream()
-> recreateContainer()
-> checkAll()
```
