#!/usr/bin/env node

import { ensureTool, ensureWebDependencies, printHelp, run, webDir } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("check-web.mjs", "Install locked web dependencies when needed, then run type check and production build.");
  process.exit(0);
}

ensureTool("node", ["--version"]);
ensureTool("npm", ["--version"]);
ensureWebDependencies();
run("npm", ["run", "check"], { cwd: webDir });
run("npm", ["run", "build"], { cwd: webDir });
