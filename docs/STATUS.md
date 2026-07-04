# Project Status

This file describes the current state of this specific project.

It must not describe the development status of the template itself.

## Current phase

Choose one:

- Not started
- Discovery
- Planning
- Implementation
- Testing
- Production
- Maintenance

Current phase: Maintenance

## Current objective

Keep private Forgejo `main` synced to latest upstream application state while preserving local agentic tooling scaffold and shipping small operator-focused improvements when needed.

## Working state

### What works

- `main` is rebased onto latest `upstream/main` and includes the private tooling scaffold in a separate follow-up commit.
- Forgejo `origin` and GitHub `upstream` remotes are configured.
- Force-push to Forgejo `origin/main` succeeded in this session.
- GitNexus index was refreshed for current repository state.
- Shared `agent-bridge` MCP broker scaffolding and docs exist for structured OpenCode <-> Caveman handoff.
- Obsolete upstream PR branches were pruned after their fixes landed upstream.
- Current upstream app state includes grouped image views, auth/login flow, i18n updates, latest docs refresh, and Docker pull rate-limit warning support.
- OpenCode MCP verification succeeded for both `gitnexus` and `agent_bridge`.
- Caveman session verification is considered successful; TASK-003 is complete.
- 12 upstream PR branches were created and pushed for TIER 1/2/3 fixes, plus 2 Forgejo-only maintenance commits.
- All Go fixes are merged into Forgejo `main` and the build succeeds.
- Application code audit found zero `FIXME`/`TODO`/`HACK` markers in runtime source files.
- Forgejo build succeeded on cleaned branch state.
- Docker pull path supports optional authenticated registry pulls via `DOCKER_REGISTRY_AUTH`.
- Web UI surfaces current Docker pull rate-limit state through a dismissible warning banner.

### What does not work yet

- FreeLLMAPI verification has not been completed.
- OpenCode environment loading has not been verified end-to-end beyond confirmed MCP connectivity.

### Unknowns

- Whether any local tooling docs still lag behind the latest upstream auth/grouped-image behavior beyond the targeted fixes already applied.
- Any next project-specific improvements beyond maintaining upstream sync and tooling verification.

## Current blockers

- No confirmed blocker.

## Environment/tooling state

- Forgejo remote: Set (`https://git.neomod.cc/neomod/docker-image-watcher.git`)
- Forgejo push auth in this session: Yes
- FreeLLMAPI verified: TBD
- OpenCode env loaded successfully: Partial (`opencode` MCP connectivity verified; full model/env route still TBD)
- GitNexus indexed: Yes
- GitNexus MCP verified in OpenCode: Yes
- Caveman Code installed/login tested: Yes
- Caveman GitNexus CLI fallback from project root: Yes
- Agent bridge MCP scaffolded: Yes
- Agent bridge MCP verified in OpenCode: Yes
- Agent bridge MCP verified in Caveman session: Yes

## Last meaningful update

- Date: 2026-07-04
- Summary: Added Docker pull rate-limit detection, authenticated pull support via `DOCKER_REGISTRY_AUTH`, a UI warning banner, and matching project documentation updates.

## Next action

TASK-005 is complete. Project remains in maintenance phase until next feature request or upstream contribution.
