# Codex / ChatGPT Coding Endpoint

Codex/ChatGPT access is separate from OpenAI API-key routing and separate from OpenCode model routing. See `docs/TOOLING_MODEL.md`.

## Access paths

### Preferred terminal path: Caveman Code

Use Caveman Code directly when it can authenticate to your ChatGPT/Codex account and you want terminal coding with lower-token behavior.

```powershell
caveman
```

Inside Caveman:

```text
/login
```

### Optional IDE path: VS Code Codex

Use Codex in VS Code when you want IDE-integrated coding.

This is optional in the template. Caveman Code may be enough for many workflows because it can provide a terminal-based Codex/ChatGPT path when configured.

## API billing separation

Do not confuse:

- ChatGPT/Codex subscription or Business allowance;
- OpenAI API key billing;
- DeepSeek API billing;
- FreeLLMAPI free/aggregated usage.

OpenAI API backup should only be used when explicitly selected and intentionally funded.

## When to use Codex/ChatGPT path

Use it for:

- heavy implementation;
- hard debugging;
- high-value refactors;
- tasks where Business/Codex allowance is preferred over personal API credit.

Use OpenCode/DeepSeek for normal lower-cost project work. Use Caveman Code when you want the separate low-token terminal/Codex-capable endpoint.


## Implementation layout

Do not scatter new source files in the repository root. Use `src/` unless `docs/ARCHITECTURE.md` defines another dedicated implementation folder. For the preferred flow, OpenCode produces the plan and Caveman Code executes the approved implementation instructions.
