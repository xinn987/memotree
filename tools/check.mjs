#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { commandName, ensureTool, ensureWebDependencies, ensureWorkspaceToolDependencies, printHelp, projectEnv, repoRoot, run, serverDir, webDir } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("check.mjs", "Run the full cross-platform MemoTree verification suite.");
  process.exit(0);
}

ensureTool("node", ["--version"]);
ensureTool("npm", ["--version"]);
ensureTool("go", ["version"]);
ensureTool("openspec", ["--version"]);

ensureWorkspaceToolDependencies();
ensureWebDependencies();

run("node", ["--test", "tools/shared.test.mjs"], { cwd: repoRoot });
run("go", ["test", "./..."], { cwd: serverDir });
run("npm", ["run", "check"], { cwd: webDir });
run("npm", ["run", "build"], { cwd: webDir });
run("openspec", ["validate", "--specs", "--strict"], { cwd: repoRoot });
for (const changeName of listActiveOpenSpecChanges()) {
  run("openspec", ["validate", changeName, "--strict"], { cwd: repoRoot });
}

function listActiveOpenSpecChanges() {
  const result = spawnSync(commandName("openspec"), ["list", "--json"], {
    cwd: repoRoot,
    env: projectEnv(),
    encoding: "utf8",
    shell: process.platform === "win32",
  });
  if (result.error) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(`Unable to list OpenSpec changes: ${result.stderr}`);
  }
  const parsed = JSON.parse(result.stdout);
  return (parsed.changes ?? []).map((change) => change.name).filter(Boolean);
}
