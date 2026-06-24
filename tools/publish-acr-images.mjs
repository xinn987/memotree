#!/usr/bin/env node

import { mkdirSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";

import { printHelp, projectEnv, repoRoot, run } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "publish-acr-images.mjs",
    "Build MemoTree runtime images, tag them for Alibaba Cloud ACR, and push them.",
    [
      "Required env:",
      "  ACR_REGISTRY    Example: registry.cn-hangzhou.aliyuncs.com",
      "  ACR_NAMESPACE   Example: memotree",
      "Optional env:",
      "  IMAGE_TAG       Defaults to current git short SHA",
      "  DOCKER_CONFIG   Defaults to .docker-config in the repository",
    ],
  );
  process.exit(0);
}

const registry = requiredEnv("ACR_REGISTRY").replace(/\/+$/, "");
const namespace = requiredEnv("ACR_NAMESPACE").replace(/^\/+|\/+$/g, "");
const imageTag = process.env.IMAGE_TAG || gitShortSHA();

const dockerConfig = process.env.DOCKER_CONFIG || path.join(repoRoot, ".docker-config");
mkdirSync(dockerConfig, { recursive: true });

const dockerEnv = { DOCKER_CONFIG: dockerConfig };
if (process.platform === "win32") {
  dockerEnv.DOCKER_HOST = "npipe:////./pipe/dockerDesktopLinuxEngine";
}
const env = projectEnv(dockerEnv);

const images = [
  { local: "memotree-api:local", remote: `${registry}/${namespace}/memotree-api:${imageTag}`, dockerfile: "deploy/api.Dockerfile" },
  { local: "memotree-worker:local", remote: `${registry}/${namespace}/memotree-worker:${imageTag}`, dockerfile: "deploy/worker.Dockerfile" },
  { local: "memotree-web:local", remote: `${registry}/${namespace}/memotree-web:${imageTag}`, dockerfile: "deploy/web.Dockerfile" },
];

for (const image of images) {
  run("docker", ["build", "-f", image.dockerfile, "-t", image.local, "-t", image.remote, "."], { cwd: repoRoot, env });
  run("docker", ["push", image.remote], { cwd: repoRoot, env });
}

console.log("\nImages pushed. Put these values in deploy/.env.staging on the server:\n");
console.log(`API_IMAGE=${images[0].remote}`);
console.log(`WORKER_IMAGE=${images[1].remote}`);
console.log(`WEB_IMAGE=${images[2].remote}`);

function requiredEnv(name) {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} is required. Run with --help for examples.`);
  }
  return value;
}

function gitShortSHA() {
  const result = spawnSync("git", ["rev-parse", "--short=12", "HEAD"], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  if (result.status === 0) {
    return result.stdout.trim();
  }
  return "manual";
}
