import { existsSync, mkdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawn, spawnSync } from "node:child_process";

// 本文件集中放置开发脚本共用逻辑，避免跨平台运行细节散落在多个脚本里。
const currentFile = fileURLToPath(import.meta.url);
const toolsDir = path.dirname(currentFile);

// 路径常量统一从 tools 目录反推，保证脚本从任意工作目录执行时都能定位到仓库根目录。
export const repoRoot = path.resolve(toolsDir, "..");
export const serverDir = path.join(repoRoot, "server");
export const webDir = path.join(repoRoot, "web");
export const composeFile = path.join(repoRoot, "deploy", "docker-compose.dev.yml");
export const localMySQLDSN = "memotree:memotree@tcp(127.0.0.1:3307)/memotree?parseTime=true";
export const localStorageEnv = {
  STORAGE_PROVIDER: "minio",
  STORAGE_ENDPOINT: "http://127.0.0.1:9000",
  STORAGE_REGION: "us-east-1",
  STORAGE_ACCESS_KEY_ID: "memotree",
  STORAGE_SECRET_ACCESS_KEY: "memotree-secret",
  STORAGE_USE_PATH_STYLE: "true",
  STORAGE_BUCKET_ORIGINALS: "memotree-originals",
  STORAGE_BUCKET_PREVIEWS: "memotree-previews",
};
export const defaultGoProxy = "https://goproxy.cn,direct";

// Windows 上 Go 可能安装后还没刷新当前终端 PATH，这里优先尝试默认安装位置。
export function commandName(command) {
  if (process.platform === "win32" && command === "go") {
    const candidates = [
      path.join(process.env.ProgramFiles ?? "C:\\Program Files", "Go", "bin", "go.exe"),
      "C:\\Program Files\\Go\\bin\\go.exe",
    ];
    const found = candidates.find((candidate) => existsSync(candidate));
    if (found) {
      return found;
    }
  }
  if (needsWindowsShell(command)) {
    return command;
  }
  if (process.platform !== "win32") {
    return command;
  }
  if (command === "npm" || command === "npx" || command === "openspec") {
    return `${command}.cmd`;
  }
  return command;
}

function needsWindowsShell(command) {
  return process.platform === "win32" && (command === "npm" || command === "npx" || command === "openspec");
}

export function projectEnv(extraEnv = {}) {
  const goCache = path.join(repoRoot, ".gocache");
  const goModCache = path.join(repoRoot, ".gomodcache");
  const goPath = path.join(repoRoot, ".gopath");
  const npmCache = path.join(webDir, ".npm-cache");

  mkdirSync(goCache, { recursive: true });
  mkdirSync(goModCache, { recursive: true });
  mkdirSync(goPath, { recursive: true });
  mkdirSync(npmCache, { recursive: true });

  return {
    ...process.env,
    // 统一把 Go 缓存放到项目内，避免不同机器的系统缓存权限差异影响测试。
    GOCACHE: goCache,
    GOMODCACHE: goModCache,
    GOPATH: goPath,
    // 手动 go run 时也走国内 Go proxy，避免默认 proxy.golang.org 在本地网络下超时。
    GOPROXY: process.env.GOPROXY || defaultGoProxy,
    // npm 会读取 npm_config_cache；这里同样收束到项目内的忽略目录。
    npm_config_cache: npmCache,
    ...extraEnv,
  };
}

export function dockerComposeArgs(args) {
  return ["compose", "-f", composeFile, ...args];
}

// 等待 compose 服务 healthcheck 通过，避免 API 在 MySQL 初始化窗口里抢跑。
export function waitForDockerServiceHealthy(service, { timeoutMs = 90_000, intervalMs = 1_000, name = service } = {}) {
  const deadline = Date.now() + timeoutMs;
  let containerID = "";
  let lastStatus = "";

  while (Date.now() < deadline) {
    const ps = spawnSync(commandName("docker"), dockerComposeArgs(["ps", "-q", service]), {
      cwd: repoRoot,
      env: projectEnv(),
      encoding: "utf8",
    });
    containerID = ps.status === 0 ? ps.stdout.trim() : "";

    if (containerID !== "") {
      const inspect = spawnSync(commandName("docker"), [
        "inspect",
        containerID,
        "--format",
        "{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}",
      ], {
        cwd: repoRoot,
        env: projectEnv(),
        encoding: "utf8",
      });
      lastStatus = inspect.status === 0 ? inspect.stdout.trim() : inspect.stderr.trim();
      if (lastStatus === "healthy" || lastStatus === "running") {
        console.log(`${name} is ready.`);
        return;
      }
      if (lastStatus === "unhealthy" || lastStatus === "exited" || lastStatus === "dead") {
        throw new Error(`${name} is ${lastStatus}. Check Docker logs for details.`);
      }
    }

    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, intervalMs);
  }

  const suffix = lastStatus ? ` Last status: ${lastStatus}.` : "";
  throw new Error(`${name} did not become healthy within ${timeoutMs}ms.${suffix}`);
}

export function run(command, args, options = {}) {
  const env = options.env ?? projectEnv();
  const cwd = options.cwd ?? repoRoot;
  const printable = [command, ...args].join(" ");
  console.log(`\n> ${printable}`);

  const result = spawnSync(commandName(command), args, {
    cwd,
    env,
    stdio: "inherit",
    shell: needsWindowsShell(command),
  });

  if (result.error) {
    const message = result.error.code === "ENOENT"
      ? `Command not found: ${command}. Please install it and make sure it is on PATH.`
      : result.error.message;
    throw new Error(message);
  }
  if (result.status !== 0) {
    throw new Error(`Command failed with exit code ${result.status}: ${printable}`);
  }
}

export function spawnWithPrefix(command, args, options = {}) {
  const env = options.env ?? projectEnv();
  const cwd = options.cwd ?? repoRoot;
  const prefix = options.prefix ?? command;

  const child = spawn(commandName(command), args, {
    cwd,
    env,
    shell: needsWindowsShell(command),
    windowsHide: true,
  });

  prefixStream(child.stdout, prefix);
  prefixStream(child.stderr, prefix);
  return child;
}

export async function waitForHTTP(url, { timeoutMs = 30_000, intervalMs = 500, name = "service", child } = {}) {
  const deadline = Date.now() + timeoutMs;
  let lastError = "";

  while (Date.now() < deadline) {
    if (child && child.exitCode !== null) {
      throw new Error(`${name} exited before becoming ready`);
    }

    try {
      const response = await fetch(url, { signal: AbortSignal.timeout(1_000) });
      if (response.ok) {
        return;
      }
      lastError = `HTTP ${response.status}`;
    } catch (error) {
      lastError = error instanceof Error ? error.message : String(error);
    }

    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }

  const suffix = lastError ? ` Last error: ${lastError}` : "";
  throw new Error(`${name} did not become ready at ${url} within ${timeoutMs}ms.${suffix}`);
}

function prefixStream(stream, prefix) {
  let buffer = "";
  stream.setEncoding("utf8");
  stream.on("data", (chunk) => {
    buffer += chunk;
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? "";
    for (const line of lines) {
      if (line.length > 0) {
        console.log(`[${prefix}] ${line}`);
      }
    }
  });
  stream.on("end", () => {
    if (buffer.length > 0) {
      console.log(`[${prefix}] ${buffer}`);
    }
  });
}

export function ensureTool(command, args = ["--version"]) {
  const result = spawnSync(commandName(command), args, {
    cwd: repoRoot,
    env: projectEnv(),
    encoding: "utf8",
    shell: needsWindowsShell(command),
  });

  if (result.error) {
    throw new Error(`Missing required tool: ${command}. Install it and rerun this script.`);
  }
  if (result.status !== 0) {
    throw new Error(`Unable to run ${command} ${args.join(" ")}.`);
  }

  const output = `${result.stdout}${result.stderr}`.trim().split(/\r?\n/)[0];
  console.log(`${command}: ${output}`);
}

export function ensureWebDependencies() {
  const nodeModules = path.join(webDir, "node_modules");
  if (existsSync(nodeModules)) {
    return;
  }
  // 使用 npm ci 保证不同机器安装到 lockfile 指定的依赖版本。
  run("npm", ["ci"], { cwd: webDir });
}

// 启动前检查固定开发端口；默认只报出占用者，显式 --kill-ports 时才清理旧进程。
export function assertPortsAvailable(ports, { kill = false } = {}) {
  const occupied = ports.flatMap((port) => findPortOwners(port));
  if (occupied.length === 0) {
    return;
  }

  if (kill) {
    killPortOwners(occupied);
    return;
  }

  console.error("\nPort conflict detected:");
  for (const owner of occupied) {
    const name = owner.name ? ` (${owner.name})` : "";
    console.error(`  port ${owner.port}: pid ${owner.pid}${name}`);
  }
  console.error("\nStop the old process, or rerun with --kill-ports if these are stale dev processes.");
  throw new Error(`Ports in use: ${[...new Set(occupied.map((owner) => owner.port))].join(", ")}`);
}

export function findPortOwners(port) {
  return process.platform === "win32" ? findWindowsPortOwners(port) : findUnixPortOwners(port);
}

// Windows 使用 netstat 找监听端口，再用 Get-Process 补充进程名，便于判断能否关闭。
function findWindowsPortOwners(port) {
  const result = spawnSync("netstat", ["-ano"], { encoding: "utf8" });
  if (result.status !== 0) {
    return [];
  }

  const pids = new Set();
  for (const line of result.stdout.split(/\r?\n/)) {
    const parts = line.trim().split(/\s+/);
    if (parts.length < 5 || parts[0] !== "TCP") {
      continue;
    }
    const localAddress = parts[1];
    const state = parts[3];
    const pid = parts[4];
    if (state === "LISTENING" && localAddress.endsWith(`:${port}`)) {
      pids.add(pid);
    }
  }

  return [...pids].map((pid) => ({ port, pid, name: getProcessName(pid) }));
}

function findUnixPortOwners(port) {
  const result = spawnSync("lsof", [`-iTCP:${port}`, "-sTCP:LISTEN", "-Pn"], { encoding: "utf8" });
  if (result.status !== 0) {
    return [];
  }

  const owners = [];
  const lines = result.stdout.trim().split(/\r?\n/).slice(1);
  for (const line of lines) {
    const parts = line.trim().split(/\s+/);
    if (parts.length >= 2) {
      owners.push({ port, pid: parts[1], name: parts[0] });
    }
  }
  return owners;
}

function getProcessName(pid) {
  if (process.platform !== "win32") {
    return "";
  }
  const command = `try { (Get-Process -Id ${pid}).ProcessName } catch { '' }`;
  const result = spawnSync("powershell", ["-NoProfile", "-Command", command], { encoding: "utf8" });
  return result.status === 0 ? result.stdout.trim() : "";
}

function killPortOwners(owners) {
  const uniquePids = [...new Set(owners.map((owner) => owner.pid))];
  for (const pid of uniquePids) {
    const ownersForPid = owners.filter((owner) => owner.pid === pid);
    const ports = ownersForPid.map((owner) => owner.port).join(", ");
    const name = ownersForPid.find((owner) => owner.name)?.name ?? "";
    console.log(`Stopping pid ${pid}${name ? ` (${name})` : ""} on port(s) ${ports}`);
    if (process.platform === "win32") {
      spawnSync("taskkill", ["/PID", String(pid), "/T", "/F"], { stdio: "inherit" });
    } else {
      spawnSync("kill", ["-TERM", String(pid)], { stdio: "inherit" });
    }
  }
}

export function printHelp(name, description, options = []) {
  const optionText = options.length > 0 ? `\n\nOptions:\n${options.map((option) => `  ${option}`).join("\n")}` : "";
  console.log(`${name}\n\n${description}\n\nUsage:\n  node tools/${name}${optionText}`);
}
