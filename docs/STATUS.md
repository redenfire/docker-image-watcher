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
- OpenCode environment loading has not been verified end-to-end beyond MCP connectivity checks.
- Caveman Code login / interactive session verification has not been completed.
- End-to-end `agent_bridge` MCP verification inside a live Caveman session has not been completed yet.

### Unknowns

- Whether a fresh interactive Caveman session will expose both `gitnexus_*` and `agent_bridge_*` tools directly, or continue to require accepted CLI fallback for GitNexus-heavy work.
- Whether any local tooling docs still lag behind the latest upstream auth/grouped-image behavior beyond the targeted fixes already applied.
- Any next project-specific improvements beyond maintaining upstream sync and tooling verification.

## Current blockers

- No repository or CI blocker confirmed. Remaining gap is interactive Caveman-session verification if full MCP parity is required.

## Environment/tooling state

- Forgejo remote: Set (`https://git.neomod.cc/neomod/docker-image-watcher.git`)
- Forgejo push auth in this session: Yes
- FreeLLMAPI verified: TBD
- OpenCode env loaded successfully: Partial (`opencode` MCP connectivity verified; full model/env route still TBD)
- GitNexus indexed: Yes
- GitNexus MCP verified in OpenCode: Yes
- Caveman Code installed/login tested: Installed; interactive login/session TBD
- Caveman GitNexus CLI fallback from project root: Yes
- Agent bridge MCP scaffolded: Yes
- Agent bridge MCP verified in OpenCode: Yes
- Agent bridge MCP verified in live Caveman session: TBD

## Last meaningful update

- Date: 2026-07-03
- Summary: Verified OpenCode MCP visibility for `gitnexus` and `agent_bridge`, installed broker dependencies locally, confirmed Forgejo build success, and narrowed remaining session verification gap to interactive Caveman MCP visibility.

## Next action

Open a fresh Caveman session from project root and inspect visible MCP tools. If `gitnexus_*` and `agent_bridge_*` are still not exposed, document CLI fallback as the current operational path.
