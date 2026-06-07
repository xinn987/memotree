#!/usr/bin/env node

import { ensureTool, printHelp, run, serverDir } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("test-server.mjs", "Run Go server and worker tests with project-local Go caches.");
  process.exit(0);
}

ensureTool("go", ["version"]);
run("go", ["test", "./..."], { cwd: serverDir });
