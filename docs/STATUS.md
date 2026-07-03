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

Keep private Forgejo `main` synced to latest upstream application state while preserving local agentic tooling scaffold and preparing project-specific verification/build work.

## Working state

### What works

- `main` is rebased onto latest `upstream/main` and includes the private tooling scaffold in a separate follow-up commit.
- Forgejo `origin` and GitHub `upstream` remotes are configured.
- Force-push to Forgejo `origin/main` succeeded in this session.
- GitNexus index was refreshed for current repository state.
- Shared `agent-bridge` MCP broker scaffolding and docs exist for structured OpenCode <-> Caveman handoff.
- Obsolete upstream PR branches were pruned after their fixes landed upstream.
- Current upstream app state includes grouped image views, auth/login flow, i18n updates, and latest docs refresh.

### What does not work yet

- FreeLLMAPI verification has not been completed.
- OpenCode environment loading has not been verified end-to-end in a live session.
- GitNexus MCP verification inside OpenCode has not been completed.
- Caveman Code install/login verification has not been completed.
- End-to-end `agent-bridge` MCP verification inside both live agent sessions has not been completed.
- Live OpenCode and Caveman endpoint verification has not been completed yet.

### Unknowns

- Whether both live agent endpoints will expose `agent_bridge` tools cleanly after local install and session restart.
- Whether any local tooling docs still lag behind the latest upstream auth/grouped-image behavior beyond the targeted fixes already applied.
- Any next project-specific improvements beyond maintaining upstream sync and tooling verification.

## Current blockers

- None confirmed at repository access level; next blocker check is Forgejo build/CI result.

## Environment/tooling state

- Forgejo remote: Set (`https://git.neomod.cc/neomod/docker-image-watcher.git`)
- Forgejo push auth in this session: Yes
- FreeLLMAPI verified: TBD
- OpenCode env loaded successfully: TBD
- GitNexus indexed: Yes
- GitNexus MCP verified in OpenCode: TBD
- Caveman Code installed/login tested: TBD
- Agent bridge MCP scaffolded: Yes
- Agent bridge MCP verified live in both endpoints: TBD

## Last meaningful update

- Date: 2026-07-03
- Summary: Synced private `main` to latest upstream, restored tracked tooling scaffold cleanly, pruned obsolete PR branches, documented private-`main` versus upstream-`pr/*` branch policy, and confirmed Forgejo build success.

## Next action

Run live OpenCode/Caveman MCP verification on cleaned `main`, then continue any endpoint-specific config cleanup surfaced by those checks.
