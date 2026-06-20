import assert from "node:assert/strict";
import { mkdirSync, writeFileSync } from "node:fs";
import { test } from "node:test";
import { tmpdir } from "node:os";
import path from "node:path";

import { resolveManagedFFmpeg } from "./shared.mjs";

test("resolveManagedFFmpeg keeps explicit FFmpeg paths", () => {
  const resolved = resolveManagedFFmpeg({
    env: {
      FFMPEG_PATH: "C:\\ffmpeg\\bin\\ffmpeg.exe",
      FFPROBE_PATH: "C:\\ffmpeg\\bin\\ffprobe.exe",
    },
  });

  assert.equal(resolved.FFMPEG_PATH, "C:\\ffmpeg\\bin\\ffmpeg.exe");
  assert.equal(resolved.FFPROBE_PATH, "C:\\ffmpeg\\bin\\ffprobe.exe");
});

test("resolveManagedFFmpeg finds project-managed npm binaries", () => {
  const root = path.join(tmpdir(), `memotree-tools-${Date.now()}`);
  const ffmpegPath = path.join(root, "node_modules", "@ffmpeg-installer", "win32-x64", "ffmpeg.exe");
  const ffprobePath = path.join(root, "node_modules", "ffprobe-static", "bin", "win32", "x64", "ffprobe.exe");
  mkdirSync(path.dirname(ffmpegPath), { recursive: true });
  mkdirSync(path.dirname(ffprobePath), { recursive: true });
  writeFileSync(ffmpegPath, "");
  writeFileSync(ffprobePath, "");

  const resolved = resolveManagedFFmpeg({
    root,
    env: {},
    platform: "win32",
    arch: "x64",
  });

  assert.equal(resolved.FFMPEG_PATH, ffmpegPath);
  assert.equal(resolved.FFPROBE_PATH, ffprobePath);
});
