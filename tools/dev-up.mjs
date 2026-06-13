#!/usr/bin/env node

import { dockerComposeArgs, ensureTool, printHelp, repoRoot, run, waitForDockerServiceHealthy, waitForHTTP } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("dev-up.mjs", "Start local Docker dependencies: MySQL and MinIO.");
  process.exit(0);
}

ensureTool("docker", ["version"]);
run("docker", dockerComposeArgs(["up", "-d", "mysql", "minio"]));
console.log("Waiting for MySQL container health check.");
waitForDockerServiceHealthy("mysql", { name: "mysql" });
console.log("Waiting for MinIO health check.");
await waitForHTTP("http://127.0.0.1:9000/minio/health/ready", { name: "minio", timeoutMs: 30_000 });
run("node", ["tools/init-storage.mjs"], { cwd: repoRoot });
