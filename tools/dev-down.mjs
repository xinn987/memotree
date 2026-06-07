#!/usr/bin/env node

import { dockerComposeArgs, ensureTool, printHelp, run } from "./shared.mjs";

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "dev-down.mjs",
    "Stop local Docker dependencies without deleting volume data.",
    ["--volumes    Also delete Docker volumes. This removes local MySQL/MinIO data."],
  );
  process.exit(0);
}

ensureTool("docker", ["version"]);

const args = ["down"];
if (process.argv.includes("--volumes")) {
  args.push("--volumes");
}
run("docker", dockerComposeArgs(args));
