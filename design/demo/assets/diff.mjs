// PC 新旧截图对比：逐页计算差异像素占比，生成 diff 高亮图。
// 用法: node design/demo/assets/diff.mjs
import sharp from "sharp";
import { readdirSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const SHOTS = join(__dirname, "shots");

const pages = ["timeline", "media-detail", "upload", "auth"];

function raw(img) {
  return sharp(img).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
}

for (const name of pages) {
  const before = join(SHOTS, `${name}-desktop-before.png`);
  const after = join(SHOTS, `${name}-desktop.png`);
  const diffOut = join(SHOTS, `${name}-desktop-diff.png`);

  const a = await raw(before);
  const b = await raw(after);

  // 尺寸不一致直接报告
  if (a.info.width !== b.info.width || a.info.height !== b.info.height) {
    console.log(`${name}: SIZE MISMATCH before=${a.info.width}x${a.info.height} after=${b.info.width}x${b.info.height}`);
    continue;
  }

  const { width, height, channels } = a.info;
  const total = width * height;
  let diffPixels = 0;
  const diff = Buffer.from(a.data); // 复制一份做差异图

  for (let i = 0; i < a.data.length; i += channels) {
    let delta = 0;
    for (let c = 0; c < 3; c++) {
      delta += Math.abs(a.data[i + c] - b.data[i + c]);
    }
    if (delta > 12) {
      // 视觉可感知的差异阈值
      diffPixels++;
      // 差异图：差异像素标红，其余灰度
      diff[i] = 255; diff[i + 1] = 0; diff[i + 2] = 0;
    } else {
      const g = Math.round((a.data[i] + a.data[i + 1] + a.data[i + 2]) / 6);
      diff[i] = g; diff[i + 1] = g; diff[i + 2] = g;
    }
  }

  const pct = ((diffPixels / total) * 100).toFixed(3);
  console.log(`${name}-desktop: ${diffPixels} / ${total} pixels differ (${pct}%)`);

  await sharp(diff, { raw: { width, height, channels } })
    .png()
    .toFile(diffOut);
}
console.log("diff images written to shots/*-desktop-diff.png");
