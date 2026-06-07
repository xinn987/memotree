#!/usr/bin/env node

import { dockerComposeArgs, ensureTool, printHelp, run } from "./shared.mjs";

const follow = process.argv.includes("--follow") || process.argv.includes("-f");

if (process.argv.includes("--help") || process.argv.includes("-h")) {
  printHelp(
    "dev-logs.mjs",
    "Show local Docker dependency logs.",
    [
      "--follow, -f    Keep following logs.",
      "mysql           Show only MySQL logs.",
      "minio           Show only MinIO logs.",
    ],
  );
  process.exit(0);
}

ensureTool("docker", ["version"]);

const services = process.argv.filter((arg) => arg === "mysql" || arg === "minio");
const args = ["logs"];
if (follow) {
  args.push("--follow");
}
args.push(...services);

run("docker", dockerComposeArgs(args));
