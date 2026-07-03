# How to Use This Template

This is the single canonical guide for starting a new project from this template. Read `docs/TOOLING_MODEL.md` first if the VS Code / OpenCode / Caveman Code / Codex roles are unclear.

## Tooling roles before setup

The template uses multiple endpoints by design:

- **VS Code** is the human editor/workbench.
- **OpenCode** is the main configurable local agent. It uses FreeLLMAPI, DeepSeek, optional OpenAI API, and GitNexus MCP.
- **Caveman Code** is a separate terminal coding agent. It can use Caveman low-token workflow and ChatGPT/Codex-style authentication when configured. It has its own Cave-side GitNexus config in `.cave/settings.json`; it does not read `opencode.json`.
- **Codex in VS Code** is optional. Use it only when you specifically want IDE-integrated Codex.
- **GitNexus** is the required repository memory/index layer.

See `docs/TOOLING_MODEL.md` for the full model.

## Canonical setup sequence

Required steps:

1. Create project folder.
2. Set Forgejo remote.
3. Deploy or verify FreeLLMAPI.
4. Set up free models in FreeLLMAPI.
5. Install/open VS Code.
6. Install/configure OpenCode.
7. Configure `.env` and launch OpenCode through the provided start script.
8. Install and index GitNexus.
9. Verify GitNexus MCP in OpenCode.
10. Install Caveman Code as the separate terminal/Codex-capable endpoint.
11. Verify Caveman Code GitNexus access or documented fallback.
12. Optionally install Caveman Skill for specific agents.
13. Configure paid/frontier OpenCode model routes.
14. Fill required project documents.
15. Start OpenCode project analysis and implementation planning.
16. Execute the approved plan with Caveman Code.

## 1. Create project folder

```powershell
mkdir my-project
cd my-project
```

If starting from this template, unpack/bootstrap first, then initialize Git.

```powershell
git init -b main
git add .
git commit -m "Initial project import"
```

## 2. Set Forgejo

Create the repository in Forgejo, then add the remote.

```powershell
git remote add origin https://git.example.com/owner/my-project.git
git push -u origin main
```

Forgejo is the canonical Git source for the project.

## 3. Deploy or verify FreeLLMAPI

FreeLLMAPI is shared infrastructure, not a per-project install.

Use `deploy/freellmapi/` for Portainer deployment notes.

Verify:

```powershell
.\scripts\check-freellmapi.ps1
```

or:

```bash
./scripts/check-freellmapi.sh
```

## 4. Set up free models in FreeLLMAPI

Configure FreeLLMAPI providers and confirm its OpenAI-compatible endpoint is reachable.

Expected project values:

```env
FREELLMAPI_BASE_URL=http://freellmapi.example.lan:3001/v1
FREELLMAPI_API_KEY=your-key
```

## 5. Install/open VS Code

VS Code is the editor shell and human workbench. It does not replace OpenCode, GitNexus, or Caveman Code.

Codex in VS Code is optional. Caveman Code is the preferred terminal path when you want Caveman low-token behavior plus ChatGPT/Codex-style auth without the VS Code Codex UI.

## 6. Install/configure OpenCode

Install OpenCode once per workstation/user. Do not vendor it into each project.

OpenCode project behavior is controlled by:

- `opencode.json`
- `.opencode/prompts/*.md`
- `AGENTS.md`
- GitNexus MCP

See `docs/OPENCODE.md`.

## 7. Configure `.env` and launch OpenCode correctly

A `.env` file is not automatically loaded by Windows, Linux, macOS, PowerShell, Bash, or OpenCode.

Create `.env`:

```powershell
copy .env.example .env
notepad .env
```

Linux/macOS:

```bash
cp .env.example .env
nano .env
```

Launch OpenCode through the wrapper so env values are loaded.

Windows:

```powershell
.\scripts\start-opencode.ps1
```

Linux/macOS:

```bash
chmod +x scripts/*.sh
./scripts/start-opencode.sh
```

Verify `opencode debug config` shows non-empty values for:

- `FREELLMAPI_BASE_URL`
- `FREELLMAPI_API_KEY`
- `DEEPSEEK_API_KEY`

If `apiKey` or `baseURL` is empty, authentication will fail.

## 8. Install and index GitNexus

GitNexus is required. It is the repository memory/index layer.

Install:

```powershell
npm install -g gitnexus@latest
```

Windows tested path when native parser packages fail:

```powershell
npm uninstall -g gitnexus
npm cache verify
$env:GITNEXUS_SKIP_OPTIONAL_GRAMMARS = "1"
npm install -g gitnexus@latest
npm install -g tree-sitter-dart tree-sitter-swift
```

Index from the repo root:

```powershell
gitnexus analyze
gitnexus status
```

During early setup before Git history exists, use:

```powershell
gitnexus analyze --skip-git
gitnexus status
```

See `docs/GITNEXUS.md`.

## 9. Verify GitNexus MCP in OpenCode

MCP is required. GitNexus MCP must work before serious OpenCode agent work.

```powershell
opencode mcp list
opencode mcp debug gitnexus
```

Inside OpenCode, ask:

```text
Use GitNexus tools to summarize this repository.
```

Do not assume MCP works just because it exists in `opencode.json`.

## 10. Install Caveman Code

Caveman Code is a separate terminal coding endpoint. It is not an OpenCode plugin and it is not the same thing as the Caveman Skill. Use it when you want terminal coding with Caveman behavior and Codex/ChatGPT authentication when configured.

Install once per workstation/user:

```powershell
npm install -g @juliusbrussee/caveman-code
caveman --version
caveman-code --version
```

On Windows, if install fails with `better-sqlite3` / `node-gyp`, uninstall Node Current/Node 26 and install Node.js LTS, then retry.

Login:

```powershell
caveman
```

Inside Caveman:

```text
/login
```

Follow the interactive prompt.

## 11. Verify Caveman Code GitNexus access

Caveman Code does not read `opencode.json`. OpenCode MCP success does not automatically give GitNexus tools to Cave/Caveman sessions.

This template includes Cave-side project config:

```text
.cave/settings.json
```

Before using Caveman Code for implementation, start Caveman Code from the project root and ask it to verify GitNexus access:

```text
Check whether GitNexus MCP tools are available in this session. If they are not visible, check whether `gitnexus status` works from the project root. Do not proceed with broad repository work until one of those checks succeeds.
```

Expected result:

- preferred: Cave exposes GitNexus MCP tools;
- acceptable with user approval: Cave can run GitNexus CLI from the project root;
- blocker: neither MCP tools nor CLI are available.

See `docs/CAVEMAN_GITNEXUS.md`.

## 12. Optionally install Caveman Skill

Caveman Skill is separate from Caveman Code.

Use it only if you want Caveman brevity behavior inside another agent such as Codex, Claude Code, Gemini, or OpenCode.

Example:

```powershell
irm https://raw.githubusercontent.com/JuliusBrussee/caveman/main/install.ps1 | iex
```

or targeted installer if supported:

```powershell
npx -y github:JuliusBrussee/caveman -- --only codex
```

Restart the target editor/agent after installing.

## 13. Configure paid/frontier OpenCode model routes

These routes are for OpenCode only. They do not control Caveman Code and they do not control Codex in VS Code.

Default OpenCode route:

```json
"model": "deepseek/deepseek-chat",
"small_model": "freellmapi/auto"
```

Meaning:

- DeepSeek is the normal cheap frontier route.
- FreeLLMAPI is the small/light/cheap route.
- OpenAI API is backup/emergency only.
- Codex/ChatGPT subscription usage should normally go through Caveman Code or VS Code Codex, not OpenCode API routing.

OpenAI API is selected only when:

1. `OPENAI_API_KEY` is loaded into the environment.
2. `model` or `small_model` points to `openai/...`, or `/models` selects an OpenAI model.
3. The user explicitly approves OpenAI API spending.

TAB in OpenCode cycles agents, not providers/models. Use `/models` to switch model route.

## 14. Fill required project documents

Before implementation, fill:

- `memory/PROJECT_BRIEF.md`
- `memory/CONSTRAINTS.md`
- `docs/STATUS.md`
- `docs/ROADMAP.md`
- `docs/TASKS.md`
- `docs/ARCHITECTURE.md`

These files must describe the actual project, not the template itself.

## 15. Start OpenCode project analysis and implementation planning

OpenCode is the planning and repository-analysis endpoint. It must use GitNexus MCP to understand the project before implementation instructions are given to Caveman Code.

See `docs/IMPLEMENTATION_WORKFLOW.md` for the complete handoff procedure.

Start with the OpenCode planning agent and use this prompt:

```text
Read the project instructions and required project documents:

- AGENTS.md
- docs/HOW_TO_USE.md
- docs/TOOLING_MODEL.md
- docs/IMPLEMENTATION_WORKFLOW.md
- docs/STATUS.md
- docs/TASKS.md
- docs/ROADMAP.md
- docs/ARCHITECTURE.md
- memory/PROJECT_BRIEF.md
- memory/CONSTRAINTS.md

Use GitNexus MCP for repository navigation and codebase understanding.

Enter plan mode.

Analyze the current project state and prepare a precise execution plan for TASK-001.

Do not edit files yet.

Your output must include:

1. task interpretation;
2. project-readiness issues, if any;
3. expected implementation folder, usually src/ unless ARCHITECTURE.md says otherwise;
4. files likely involved;
5. implementation steps;
6. checks/tests to run;
7. risks and assumptions;
8. exact instructions suitable to paste into Caveman Code for implementation.

Keep the plan scoped to TASK-001.
Do not introduce unrelated refactors.
Do not start implementation.
```

Review the OpenCode plan manually. OpenCode should improve the project plan and produce implementation instructions, not start coding during this step.

## 16. Execute approved implementation with Caveman Code

After approving the OpenCode plan, use Caveman Code as the focused coding endpoint.

Paste the approved implementation instructions into Caveman Code with this wrapper:

```text
Read AGENTS.md, docs/TOOLING_MODEL.md, docs/IMPLEMENTATION_WORKFLOW.md, docs/CAVEMAN_GITNEXUS.md, and the approved OpenCode execution plan below.

Before implementation, verify that GitNexus is available in this Caveman Code session. Prefer GitNexus MCP tools. If GitNexus MCP tools are not visible, check whether `gitnexus status` works from the project root and explicitly report that you are using CLI fallback. If neither works, stop and report the blocker.

Implement only the approved TASK-001 scope.

Do not perform unrelated refactors.
Do not modify secrets.
Do not modify generated indexes.
Do not scatter new files in the repository root.
Place new implementation files under the dedicated implementation folder identified by the plan, usually src/ unless docs/ARCHITECTURE.md says otherwise.
Use the smallest correct change.

After implementation:

1. state whether GitNexus MCP or GitNexus CLI fallback was used;
2. list changed files;
3. run the requested checks;
4. report check results;
5. explain any deviations from the plan;
6. update docs/STATUS.md only if the project state changed.

Approved OpenCode plan:

[PASTE OPENCODE PLAN HERE]
```

Then:

1. Review changed files.
2. Verify new code is in the expected implementation folder.
3. Run checks again if needed.
4. Update status/docs if project state changed.
5. Commit one coherent task-sized change.
