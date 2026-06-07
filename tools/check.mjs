#!/usr/bin/env node

import { ensureTool, ensureWebDependencies, printHelp, repoRoot, run, serverDir, webDir } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("check.mjs", "Run the full cross-platform MemoTree verification suite.");
  process.exit(0);
}

ensureTool("node", ["--version"]);
ensureTool("npm", ["--version"]);
ensureTool("go", ["version"]);
ensureTool("openspec", ["--version"]);

ensureWebDependencies();

run("go", ["test", "./..."], { cwd: serverDir });
run("npm", ["run", "check"], { cwd: webDir });
run("npm", ["run", "build"], { cwd: webDir });
run("openspec", ["validate", "family-shared-album-mvp", "--strict"], { cwd: repoRoot });
