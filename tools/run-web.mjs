#!/usr/bin/env node

import { assertPortsAvailable, ensureTool, ensureWebDependencies, printHelp, run, webDir } from "./shared.mjs";

const killPorts = process.argv.includes("--kill-ports");

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "run-web.mjs",
    "Run the Vite web development server.",
    ["--kill-ports  Stop processes occupying the web dev port before startup."],
  );
  process.exit(0);
}

ensureTool("node", ["--version"]);
ensureTool("npm", ["--version"]);
try {
  assertPortsAvailable([5173], { kill: killPorts });
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
ensureWebDependencies();
run("npm", ["run", "dev"], { cwd: webDir });
