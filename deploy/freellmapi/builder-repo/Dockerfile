# Optional private-mirror FreeLLMAPI container image.
# Prefer upstream official image unless you need an internal Forgejo mirror:
#   ghcr.io/tashfeenahmed/freellmapi:latest

FROM node:20-bookworm-slim AS builder

WORKDIR /app

# better-sqlite3 may need native compilation when no prebuilt binary exists.
# Keep build tools in the builder stage only.
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    python3 \
    make \
    g++ \
  && rm -rf /var/lib/apt/lists/*

# FreeLLMAPI is an npm workspace project. Copy full source before npm ci so npm
# can see shared/server/client package manifests and install workspace deps.
COPY . .
RUN npm ci --include=dev
RUN npm run build
RUN npm prune --omit=dev

FROM node:20-bookworm-slim AS runtime

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3001

COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules

COPY --from=builder /app/shared ./shared
COPY --from=builder /app/server/package*.json ./server/
COPY --from=builder /app/server/node_modules ./server/node_modules
COPY --from=builder /app/server/dist ./server/dist

COPY --from=builder /app/client/package*.json ./client/
COPY --from=builder /app/client/dist ./client/dist

EXPOSE 3001

CMD ["npm", "run", "start", "-w", "server"]
