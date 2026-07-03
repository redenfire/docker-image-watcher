# New Project Memo

Use this as the quick checklist when creating a project from the template.

1. Create project folder.
2. Initialize Git and set Forgejo remote.
3. Copy `.env.example` to `.env` and fill values.
4. Verify FreeLLMAPI.
5. Install/start OpenCode through `scripts/start-opencode.*`.
6. Install/index GitNexus.
7. Verify `opencode mcp debug gitnexus`.
8. Install/login Caveman Code.
9. Configure paid/frontier model routes.
10. Fill baseline project docs.
11. Start with OpenCode plan agent.

Required docs to fill:

- `memory/PROJECT_BRIEF.md`
- `memory/CONSTRAINTS.md`
- `docs/STATUS.md`
- `docs/ROADMAP.md`
- `docs/TASKS.md`
- `docs/ARCHITECTURE.md`

First planning prompt:

```text
Read AGENTS.md, docs/HOW_TO_USE.md, docs/STATUS.md, docs/TASKS.md, docs/ROADMAP.md, docs/ARCHITECTURE.md, memory/PROJECT_BRIEF.md, and memory/CONSTRAINTS.md. Use GitNexus MCP. Plan TASK-001 only. Do not edit files.
```
