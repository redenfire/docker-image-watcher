#!/usr/bin/env node
import { randomUUID } from "node:crypto";
import { existsSync } from "node:fs";
import { mkdir, readFile, readdir, rename, rm, stat, unlink, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { CallToolRequestSchema, ListToolsRequestSchema } from "@modelcontextprotocol/sdk/types.js";

const SCRIPT_DIR = dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = resolve(SCRIPT_DIR, "..", "..");
const ROOT_DIR = join(REPO_ROOT, "tmp", "agent-bridge");
const QUEUE_DIR = join(ROOT_DIR, "queue");
const ARCHIVE_DIR = join(ROOT_DIR, "archive");
const LOCK_DIR = join(ROOT_DIR, "locks");
const HANDOFF_ID_PATTERN = /^handoff-[A-Za-z0-9_-]+$/;
const SUPPORTED_AGENTS = new Set(["caveman", "opencode"]);
const ACTIVE_STATUSES = new Set(["queued", "claimed", "in_progress"]);
const FINALIZABLE_STATUSES = new Set(["claimed", "in_progress"]);
const LOCK_TTL_MS = 5 * 60 * 1000;

await ensureDirectories();

const server = new Server(
  { name: "agent_bridge", version: "1.0.0" },
  { capabilities: { tools: {} } },
);

server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    {
      name: "handoff_create",
      description: "Create a new handoff task for another agent to execute.",
      inputSchema: {
        type: "object",
        properties: {
          target_agent: { type: "string", description: "caveman or opencode" },
          title: { type: "string" },
          goal: { type: "string" },
          files: { type: "array", items: { type: "string" } },
          steps: { type: "array", items: { type: "string" } },
          verification: { type: "array", items: { type: "string" } },
          constraints: { type: "array", items: { type: "string" } },
          base_branch: { type: "string", default: "main" },
          create_pr_branches: { type: "boolean", default: false },
          metadata: { type: "object", additionalProperties: true },
        },
        required: ["target_agent", "title"],
      },
    },
    {
      name: "handoff_list",
      description: "List pending handoff tasks for a target agent.",
      inputSchema: {
        type: "object",
        properties: {
          target_agent: { type: "string" },
          status: { type: "string", enum: ["queued", "claimed", "in_progress"] },
          limit: { type: "number", default: 10 },
        },
        required: ["target_agent"],
      },
    },
    {
      name: "handoff_claim",
      description: "Claim a queued handoff task.",
      inputSchema: {
        type: "object",
        properties: {
          handoff_id: { type: "string" },
          agent: { type: "string" },
        },
        required: ["handoff_id", "agent"],
      },
    },
    {
      name: "handoff_get",
      description: "Get full handoff details and result.",
      inputSchema: {
        type: "object",
        properties: {
          handoff_id: { type: "string" },
        },
        required: ["handoff_id"],
      },
    },
    {
      name: "handoff_complete",
      description: "Mark a handoff as completed with results.",
      inputSchema: {
        type: "object",
        properties: {
          handoff_id: { type: "string" },
          agent: { type: "string" },
          result: {
            type: "object",
            properties: {
              status: { type: "string", enum: ["success"] },
              changed_files: { type: "array", items: { type: "string" } },
              checks: { type: "array", items: { type: "object", additionalProperties: true } },
              commits: { type: "array", items: { type: "string" } },
              branches: { type: "array", items: { type: "string" } },
              pr_urls: { type: "array", items: { type: "string" } },
              deviations: { type: "array", items: { type: "string" } },
            },
            additionalProperties: true,
          },
        },
        required: ["handoff_id", "agent", "result"],
      },
    },
    {
      name: "handoff_fail",
      description: "Mark a handoff as failed.",
      inputSchema: {
        type: "object",
        properties: {
          handoff_id: { type: "string" },
          agent: { type: "string" },
          error: {
            type: "object",
            properties: {
              summary: { type: "string" },
              details: { type: "string" },
            },
            required: ["summary"],
            additionalProperties: true,
          },
        },
        required: ["handoff_id", "agent", "error"],
      },
    },
  ],
}));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name } = request.params;
  const args = request.params.arguments ?? {};

  try {
    switch (name) {
      case "handoff_create":
        return text(await createHandoff(args));
      case "handoff_list":
        return text(await listHandoffs(args));
      case "handoff_claim":
        return text(await claimHandoff(args));
      case "handoff_get":
        return text(await getHandoff(args));
      case "handoff_complete":
        return text(await completeHandoff(args));
      case "handoff_fail":
        return text(await failHandoff(args));
      default:
        throw new Error(`Unknown tool: ${name}`);
    }
  } catch (error) {
    return errorText(name, error);
  }
});

const transport = new StdioServerTransport();
await server.connect(transport);

async function createHandoff(args) {
  const handoffId = `handoff-${Date.now()}-${randomUUID().slice(0, 8)}`;
  const now = new Date().toISOString();
  const task = {
    handoff_id: handoffId,
    status: "queued",
    created_at: now,
    updated_at: now,
    target_agent: requireAgentArg(args, "target_agent"),
    title: requireStringArg(args, "title"),
    goal: asString(args.goal),
    files: asStringArray(args.files),
    steps: asStringArray(args.steps),
    verification: asStringArray(args.verification),
    constraints: asStringArray(args.constraints),
    base_branch: asString(args.base_branch) || "main",
    create_pr_branches: Boolean(args.create_pr_branches),
    metadata: optionalObjectArg(args, "metadata"),
    log: [
      {
        at: now,
        event: "created",
      },
    ],
  };

  await writeJson(queuePathFor(handoffId), task);
  return { handoff_id: handoffId, status: task.status };
}

async function listHandoffs(args) {
  const targetAgent = requireAgentArg(args, "target_agent");
  const requestedStatus = args.status ? String(args.status) : null;
  const limit = normalizeLimit(args.limit);
  const entries = [];

  for (const fileName of await readdir(QUEUE_DIR)) {
    if (!fileName.endsWith(".json")) {
      continue;
    }

    let task;
    try {
      task = await readJson(join(QUEUE_DIR, fileName));
    } catch (error) {
      if (error && typeof error === "object" && "code" in error && error.code === "ENOENT") {
        continue;
      }
      throw error;
    }

    if (existsSync(archivePathFor(task.handoff_id))) {
      continue;
    }
    if (task.target_agent !== targetAgent) {
      continue;
    }
    if (requestedStatus && task.status !== requestedStatus) {
      continue;
    }
    if (!requestedStatus && !ACTIVE_STATUSES.has(task.status)) {
      continue;
    }

    entries.push({
      handoff_id: task.handoff_id,
      title: task.title,
      status: task.status,
      created_at: task.created_at,
      claimed_by: task.claimed_by ?? null,
    });
  }

  entries.sort((left, right) => left.created_at.localeCompare(right.created_at));
  return entries.slice(0, limit);
}

async function claimHandoff(args) {
  const handoffId = requireHandoffId(args.handoff_id);
  const agent = requireAgentArg(args, "agent");

  return withHandoffLock(handoffId, async () => {
    const task = await loadActiveHandoff(handoffId);

    if (task.target_agent !== agent) {
      throw new Error(`Handoff targets ${task.target_agent}, not ${agent}`);
    }
    if (task.status !== "queued") {
      throw new Error(`Handoff is not claimable: ${task.status}`);
    }

    const now = new Date().toISOString();
    task.status = "claimed";
    task.claimed_by = agent;
    task.claimed_at = now;
    task.updated_at = now;
    task.log.push({ at: now, agent, event: "claimed" });
    await writeJson(queuePathFor(handoffId), task);

    return {
      handoff_id: handoffId,
      status: task.status,
      claimed_by: agent,
    };
  });
}

async function getHandoff(args) {
  const handoffId = requireHandoffId(args.handoff_id);
  const archivePath = archivePathFor(handoffId);
  if (existsSync(archivePath)) {
    return await readJson(archivePath);
  }

  const queuePath = queuePathFor(handoffId);
  if (existsSync(queuePath)) {
    return await readJson(queuePath);
  }

  throw new Error(`Handoff not found: ${handoffId}`);
}

async function completeHandoff(args) {
  const handoffId = requireHandoffId(args.handoff_id);
  const agent = requireAgentArg(args, "agent");
  const result = requireObjectArg(args, "result");

  if (!requireStringValue(result.status, "result.status") || result.status !== "success") {
    throw new Error("result.status must be 'success'");
  }

  return withHandoffLock(handoffId, async () => {
    const task = await loadActiveHandoff(handoffId);
    assertFinalizableBy(task, agent);

    const now = new Date().toISOString();
    task.status = "completed";
    task.completed_at = now;
    task.updated_at = now;
    task.result = result;
    task.log.push({ at: now, agent, event: "completed" });

    await archiveHandoff(task);
    await unlink(queuePathFor(handoffId)).catch(() => {});

    return {
      handoff_id: handoffId,
      status: task.status,
      completed_at: now,
    };
  });
}

async function failHandoff(args) {
  const handoffId = requireHandoffId(args.handoff_id);
  const agent = requireAgentArg(args, "agent");
  const errorPayload = requireObjectArg(args, "error");
  requireStringValue(errorPayload.summary, "error.summary");

  return withHandoffLock(handoffId, async () => {
    const task = await loadActiveHandoff(handoffId);
    assertFinalizableBy(task, agent);

    const now = new Date().toISOString();
    task.status = "failed";
    task.failed_at = now;
    task.updated_at = now;
    task.error = errorPayload;
    task.log.push({
      at: now,
      agent,
      event: "failed",
      summary: task.error.summary ?? "",
    });

    await archiveHandoff(task);
    await unlink(queuePathFor(handoffId)).catch(() => {});

    return {
      handoff_id: handoffId,
      status: task.status,
      failed_at: now,
    };
  });
}

async function loadActiveHandoff(handoffId) {
  if (existsSync(archivePathFor(handoffId))) {
    throw new Error(`Handoff already finalized: ${handoffId}`);
  }

  const queuePath = queuePathFor(handoffId);
  if (!existsSync(queuePath)) {
    throw new Error(`Handoff not found in active queue: ${handoffId}`);
  }
  return readJson(queuePath);
}

async function archiveHandoff(task) {
  await writeJson(archivePathFor(task.handoff_id), task);
}

function queuePathFor(handoffId) {
  return join(QUEUE_DIR, `${requireHandoffId(handoffId)}.json`);
}

function archivePathFor(handoffId) {
  return join(ARCHIVE_DIR, `${requireHandoffId(handoffId)}.json`);
}

function requireHandoffId(value) {
  const handoffId = String(value);
  if (!HANDOFF_ID_PATTERN.test(handoffId)) {
    throw new Error(`Invalid handoff_id: ${handoffId}`);
  }
  return handoffId;
}

function assertFinalizableBy(task, agent) {
  if (task.target_agent !== agent) {
    throw new Error(`Handoff targets ${task.target_agent}, not ${agent}`);
  }
  if (!FINALIZABLE_STATUSES.has(task.status)) {
    throw new Error(`Handoff is not ready for finalization: ${task.status}`);
  }
  if (!task.claimed_by) {
    throw new Error("Handoff has not been claimed");
  }
  if (task.claimed_by !== agent) {
    throw new Error(`Handoff claimed by ${task.claimed_by}, not ${agent}`);
  }
}

function normalizeLimit(value) {
  const limit = Number(value ?? 10);
  if (!Number.isFinite(limit) || limit < 1) {
    return 10;
  }
  return Math.min(Math.trunc(limit), 100);
}

function asString(value) {
  if (typeof value !== "string") {
    return "";
  }
  return value;
}

function asStringArray(value) {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.map((entry) => String(entry));
}

function optionalObjectArg(args, key) {
  const value = args[key];
  if (value === undefined) {
    return {};
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${key} must be an object`);
  }
  return value;
}

function requireObjectArg(args, key) {
  const value = args[key];
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${key} is required and must be an object`);
  }
  return value;
}

function requireStringArg(args, key) {
  return requireStringValue(args[key], key);
}

function requireAgentArg(args, key) {
  const agent = requireStringValue(args[key], key);
  if (!SUPPORTED_AGENTS.has(agent)) {
    throw new Error(`${key} must be one of: ${Array.from(SUPPORTED_AGENTS).join(", ")}`);
  }
  return agent;
}

function requireStringValue(value, key) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${key} is required and must be a non-empty string`);
  }
  return value;
}

async function readJson(filePath) {
  const content = await readFile(filePath, "utf8");
  return JSON.parse(content);
}

async function writeJson(filePath, data) {
  const tempPath = `${filePath}.${process.pid}.${randomUUID().slice(0, 8)}.tmp`;
  try {
    await writeFile(tempPath, JSON.stringify(data, null, 2) + "\n", "utf8");
    await rename(tempPath, filePath);
  } catch (error) {
    await rm(tempPath, { force: true }).catch(() => {});
    throw error;
  }
}

async function ensureDirectories() {
  await mkdir(QUEUE_DIR, { recursive: true });
  await mkdir(ARCHIVE_DIR, { recursive: true });
  await mkdir(LOCK_DIR, { recursive: true });
}

async function withHandoffLock(handoffId, callback) {
  const lockPath = join(LOCK_DIR, `${requireHandoffId(handoffId)}.lock`);
  await acquireLock(lockPath, handoffId);

  try {
    return await callback();
  } finally {
    await rm(lockPath, { recursive: true, force: true });
  }
}

async function acquireLock(lockPath, handoffId) {
  try {
    await mkdir(lockPath);
    return;
  } catch (error) {
    if (!error || typeof error !== "object" || !("code" in error) || error.code !== "EEXIST") {
      throw error;
    }
  }

  const lockAge = await getLockAgeMs(lockPath);
  if (lockAge !== null && lockAge > LOCK_TTL_MS) {
    await rm(lockPath, { recursive: true, force: true });
    await mkdir(lockPath);
    return;
  }

  throw new Error(`Handoff is locked: ${handoffId}`);
}

async function getLockAgeMs(lockPath) {
  try {
    const lockStat = await stat(lockPath);
    return Date.now() - lockStat.mtimeMs;
  } catch (error) {
    if (error && typeof error === "object" && "code" in error && error.code === "ENOENT") {
      return null;
    }
    throw error;
  }
}

function text(payload) {
  return {
    content: [
      {
        type: "text",
        text: JSON.stringify(payload, null, 2),
      },
    ],
  };
}

function errorText(toolName, error) {
  return {
    content: [
      {
        type: "text",
        text: JSON.stringify(
          {
            error: error instanceof Error ? error.message : String(error),
            tool: toolName,
          },
          null,
          2,
        ),
      },
    ],
    isError: true,
  };
}
