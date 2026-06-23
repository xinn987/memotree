#!/usr/bin/env node

import { mkdirSync } from "node:fs";
import path from "node:path";

import { printHelp, projectEnv, repoRoot, run } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "build-images.mjs",
    "Build local MemoTree API, Worker, and Web runtime images.",
    ["--help  Show this help message."],
  );
  process.exit(0);
}

const dockerConfigDir = path.join(repoRoot, ".docker-config");
mkdirSync(dockerConfigDir, { recursive: true });
const dockerEnv = { DOCKER_CONFIG: dockerConfigDir };
if (process.platform === "win32") {
  dockerEnv.DOCKER_HOST = "npipe:////./pipe/dockerDesktopLinuxEngine";
}
const env = projectEnv(dockerEnv);

const images = [
  ["memotree-api:local", "deploy/api.Dockerfile"],
  ["memotree-worker:local", "deploy/worker.Dockerfile"],
  ["memotree-web:local", "deploy/web.Dockerfile"],
];

for (const [tag, dockerfile] of images) {
  run("docker", ["build", "-f", dockerfile, "-t", tag, "."], { cwd: repoRoot, env });
}
