#!/usr/bin/env node

import { ensureTool, localStorageEnv, printHelp, projectEnv, run, serverDir } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "init-storage.mjs",
    "Idempotently create local S3-compatible storage buckets for MemoTree.",
    ["--help  Show this help message."],
  );
  process.exit(0);
}

ensureTool("go", ["version"]);

const env = projectEnv(withLocalStorageDefaults());
run("go", ["run", "./devtools/cmd/init-storage"], { cwd: serverDir, env });

function withLocalStorageDefaults() {
  return Object.fromEntries(
    Object.entries(localStorageEnv).map(([key, value]) => [key, process.env[key] || value]),
  );
}
