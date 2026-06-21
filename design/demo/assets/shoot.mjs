// 截图自检脚本 —— 用 playwright 把四个 demo 页按桌面 + 移动视口截图。
// 用法: node design/demo/assets/shoot.mjs
import { chromium } from "playwright";
import { mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT = join(__dirname, "shots");
mkdirSync(OUT, { recursive: true });

const BASE = "http://localhost:8401/demo/";
const pages = [
  ["timeline", "timeline.html"],
  ["media-detail", "media-detail.html"],
  ["upload", "upload.html"],
  ["auth", "auth.html"],
];
const viewports = [
  ["desktop", 1440, 900],
  ["mobile", 390, 844],
];

const browser = await chromium.launch();
for (const [vname, w, h] of viewports) {
  const ctx = await browser.newContext({
    viewport: { width: w, height: h },
    deviceScaleFactor: 2,
  });
  const page = await ctx.newPage();
  for (const [name, file] of pages) {
    await page.goto(BASE + file, { waitUntil: "networkidle", timeout: 30000 });
    // 等字体和图片就位
    await page.waitForTimeout(1500);
    // 部分页（timeline/auth）较高，截全页
    const out = join(OUT, `${name}-${vname}.png`);
    await page.screenshot({ path: out, fullPage: true });
    console.log("shot:", out);
  }
  await ctx.close();
}
await browser.close();
console.log("done");
