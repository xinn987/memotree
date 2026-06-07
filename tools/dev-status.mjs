#!/usr/bin/env node

import { dockerComposeArgs, ensureTool, printHelp, run } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp("dev-status.mjs", "Show local Docker dependency container status.");
  process.exit(0);
}

ensureTool("docker", ["version"]);
run("docker", dockerComposeArgs(["ps"]));
