# Security Notes

## Secrets

Do not commit real secrets.

Files that may contain secrets:

- `.env`
- Portainer stack env files
- provider API keys
- FreeLLMAPI encryption key

Commit examples only:

- `.env.example`
- `deploy/freellmapi/portainer/stack.env.example`

## OpenCode and API spend

OpenAI API backup use must be explicit.

The presence of an OpenAI provider block does not authorize spending. OpenAI API is selected only when an OpenAI model is selected and `OPENAI_API_KEY` is loaded.

## FreeLLMAPI encryption key

`ENCRYPTION_KEY` must be generated once and kept permanently.

Changing it after database creation may make stored provider credentials unusable.

## GitNexus

Do not commit generated indexes or local cache data.

## Agent safety

Agents must not modify secrets, local credentials, generated indexes, or deployment keys unless explicitly instructed.
