#!/usr/bin/env node

import {
  assertPortsAvailable,
  dockerComposeArgs,
  ensureTool,
  ensureWebDependencies,
  localStorageEnv,
  localMySQLDSN,
  printHelp,
  projectEnv,
  repoRoot,
  run,
  spawnWithPrefix,
  waitForDockerServiceHealthy,
  waitForHTTP,
} from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "dev.mjs",
    "Start the local development environment: Docker dependencies, API, and web app.",
    [
      "--memory      Run API with in-memory store instead of local Docker MySQL.",
      "--no-deps     Do not start Docker dependencies before API/web.",
      "--kill-ports  Stop processes occupying API/web dev ports before startup.",
    ],
  );
  process.exit(0);
}

const useMemoryStore = process.argv.includes("--memory");
const skipDeps = process.argv.includes("--no-deps");
const killPorts = process.argv.includes("--kill-ports");

ensureTool("node", ["--version"]);
ensureTool("npm", ["--version"]);
ensureTool("go", ["version"]);
try {
  assertPortsAvailable([8080, 5173], { kill: killPorts });
} catch (error) {
  console.error(error.message);
  process.exit(1);
}

if (!skipDeps && !useMemoryStore) {
  ensureTool("docker", ["version"]);
  run("docker", dockerComposeArgs(["up", "-d", "mysql", "minio"]));
  console.log("Waiting for MySQL container health check.");
  waitForDockerServiceHealthy("mysql", { name: "mysql" });
}

ensureWebDependencies();

const apiArgs = ["tools/run-api.mjs"];
if (!useMemoryStore) {
  apiArgs.push("--mysql");
}

const apiEnv = projectEnv(useMemoryStore ? {} : {
  ...withLocalStorageDefaults(),
  MYSQL_DSN: process.env.MYSQL_DSN || localMySQLDSN,
});
let shuttingDown = false;
const children = [];

const apiChild = spawnWithPrefix(process.execPath, apiArgs, { cwd: repoRoot, env: apiEnv, prefix: "api" });
children.push(apiChild);
watchChild(apiChild);

try {
  console.log("Waiting for API health check: http://127.0.0.1:8080/healthz");
  await waitForHTTP("http://127.0.0.1:8080/healthz", { name: "api", child: apiChild });
  console.log("API is ready. Starting web dev server.");
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  shutdown(1);
  await new Promise((resolve) => setTimeout(resolve, 350));
  process.exit(1);
}

const webChild = spawnWithPrefix(process.execPath, ["tools/run-web.mjs"], { cwd: repoRoot, prefix: "web" });
children.push(webChild);
watchChild(webChild);

function watchChild(child) {
  child.on("exit", (code, signal) => {
    if (shuttingDown) {
      return;
    }
    if (code !== 0) {
      console.error(`A dev process exited with code ${code ?? "null"} signal ${signal ?? "null"}.`);
      shutdown(1);
    }
  });
}

process.on("SIGINT", () => shutdown(0));
process.on("SIGTERM", () => shutdown(0));

function shutdown(exitCode) {
  shuttingDown = true;
  for (const child of children) {
    if (!child.killed) {
      child.kill();
    }
  }
  setTimeout(() => process.exit(exitCode), 300);
}

function withLocalStorageDefaults() {
  return Object.fromEntries(
    Object.entries(localStorageEnv).map(([key, value]) => [key, process.env[key] || value]),
  );
}
