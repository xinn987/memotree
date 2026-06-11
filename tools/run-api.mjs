#!/usr/bin/env node

import { assertPortsAvailable, ensureTool, localMySQLDSN, localStorageEnv, printHelp, projectEnv, run, serverDir } from "./shared.mjs";

const useMySQL = process.argv.includes("--mysql");
const killPorts = process.argv.includes("--kill-ports");

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "run-api.mjs",
    "Run the Go API with project-local Go caches and a stable Go module proxy.",
    [
      "--mysql       Use local Docker MySQL on 127.0.0.1:3307.",
      "--kill-ports  Stop processes occupying the API dev port before startup.",
      "--help        Show this help message.",
    ],
  );
  process.exit(0);
}

ensureTool("go", ["version"]);
try {
  assertPortsAvailable([8080], { kill: killPorts });
} catch (error) {
  console.error(error.message);
  process.exit(1);
}

const extraEnv = useMySQL ? {
  ...withLocalStorageDefaults(),
  MYSQL_DSN: process.env.MYSQL_DSN || localMySQLDSN,
} : {};
const env = projectEnv(extraEnv);

if (useMySQL) {
  console.log(`MYSQL_DSN: ${env.MYSQL_DSN}`);
} else {
  console.log("MYSQL_DSN: <empty> (using in-memory store)");
}

run("go", ["run", "./api/cmd/api"], { cwd: serverDir, env });

function withLocalStorageDefaults() {
  return Object.fromEntries(
    Object.entries(localStorageEnv).map(([key, value]) => [key, process.env[key] || value]),
  );
}
