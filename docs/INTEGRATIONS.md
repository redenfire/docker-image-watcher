# Integrations

This document lists integrations. For the role of each editor/agent endpoint, read `docs/TOOLING_MODEL.md` first.

This document summarizes integrations and points to the detailed docs.

## Required integrations

| Integration | Required | Detail |
|---|---:|---|
| Forgejo | Yes | Canonical Git remote. |
| FreeLLMAPI | Yes | `docs/FREELLMAPI.md` |
| OpenCode | Yes | `docs/OPENCODE.md` |
| GitNexus | Yes | `docs/GITNEXUS.md` |
| GitNexus MCP | Yes | `docs/OPENCODE.md` and `docs/GITNEXUS.md` |
| Caveman Code | Yes | `docs/CAVEMAN.md` |

## Optional integrations

| Integration | Optional use |
|---|---|
| Caveman Skill | Install into specific agents for brevity behavior. |
| Codex in VS Code | IDE-integrated ChatGPT/Codex coding endpoint. |
| OpenAI API | Backup/emergency API route only. |

## Alignment with HOW_TO_USE

Follow `docs/HOW_TO_USE.md` for the canonical order. This file is a map, not a second setup guide.
