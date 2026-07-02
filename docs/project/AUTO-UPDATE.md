# Auto-Update

## How It Is Enabled

Auto-update can be toggled:

- from web UI toggle in `web/index.html`
- via `POST /api/images/{id}/auto-update`

Toggle state is returned as:

```json
{"auto_update":true}
```

## Persistence Model

State is stored as JSON map in `AUTO_FILE`.

Example:

```json
{
  "3d4c5b6a7f8e": true,
  "9f8e7d6c5b4a": false
}
```

Implementation points:

- `App.autoSaved` keeps in-memory map
- `loadAuto()` reads file on startup
- `saveAuto()` writes file after each toggle
- keys use short 12-character container IDs

## Check Interval

Background refresh loop in `main.go`:

- runs one initial `checkAll()` on startup
- then runs every 10 minutes via `time.NewTicker(10 * time.Minute)`

## Cooldown

Automatic updates use per-container cooldown:

- cooldown window: 5 minutes
- stored in `App.cooldowns`
- checked inside `checkAll()` before triggering auto-update

This avoids repeated rapid recreate cycles when a container stays reported as outdated.

## Trigger Conditions

Auto-update only triggers when all conditions are true:

- container appears in current running container list
- auto-update is enabled for that container
- computed status is `outdated`
- no active 5-minute cooldown blocks that container

## Concurrency and State Safety

- `sync.RWMutex` protects `images`, `autoSaved`, and `cooldowns`
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
