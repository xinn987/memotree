#!/usr/bin/env node

import { dockerComposeArgs, ensureTool, printHelp, run } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("dev-up.mjs", "Start local Docker dependencies: MySQL and MinIO.");
  process.exit(0);
}

ensureTool("docker", ["version"]);
run("docker", dockerComposeArgs(["up", "-d", "mysql", "minio"]));
