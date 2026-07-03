# Internal Memo

This file records template/tooling notes for future maintainers of this repository.

It is not the project task board and should not drive normal implementation work.

## v6.7 template decisions

- GitNexus remains required as the memory/index layer.
- OpenCode MCP must be verified before serious agent work.
- `.env` loading is explicit through start scripts on both Windows and Linux/macOS.
- Caveman Code is the main terminal low-token endpoint.
- Caveman Skill is documented separately as an optional integration for other agents.
- OpenAI API route is backup/emergency only unless explicitly selected.
- Codex/ChatGPT subscription usage is separate from OpenAI API-key billing.


## Template note

This template includes `docs/TOOLING_MODEL.md` to clarify the VS Code / OpenCode / Caveman Code / Codex endpoint split. Remove or adapt this note after the project has its own status content.

## v6.7 tooling model polish

The package explicitly separates VS Code, OpenCode, Caveman Code, Codex in VS Code, Caveman Skill, GitNexus, and model providers. This prevents confusion between editor, agent endpoint, memory layer, and model route.
