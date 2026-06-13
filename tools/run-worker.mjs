#!/usr/bin/env node

import { ensureTool, localMySQLDSN, localStorageEnv, printHelp, projectEnv, run, serverDir } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "run-worker.mjs",
    "Run the Go media worker with local MySQL and S3-compatible object storage defaults.",
    [
      "--help  Show this help message.",
    ],
  );
  process.exit(0);
}

ensureTool("go", ["version"]);

const env = projectEnv({
  ...withLocalStorageDefaults(),
  MYSQL_DSN: process.env.MYSQL_DSN || localMySQLDSN,
});

console.log(`MYSQL_DSN: ${env.MYSQL_DSN}`);
run("go", ["run", "./worker/cmd/worker"], { cwd: serverDir, env });

function withLocalStorageDefaults() {
  return Object.fromEntries(
    Object.entries(localStorageEnv).map(([key, value]) => [key, process.env[key] || value]),
  );
}
