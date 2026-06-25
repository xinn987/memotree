#!/usr/bin/env node

// 汇总本地开发常用命令，作为不需要记脚本名的入口。
const sections = [
  {
    title: "开发启动",
    commands: [
      ["node tools/dev.mjs", "一键启动本地开发环境：Docker 依赖、API、前端"],
      ["node tools/dev.mjs --kill-ports", "启动前清理占用 8080/5173 的旧开发进程"],
      ["node tools/dev.mjs --memory", "一键启动 API/前端，但 API 使用内存 store"],
      ["node tools/run-api.mjs", "启动 API（内存 store，重启后数据会丢）"],
      ["node tools/run-api.mjs --mysql", "启动 API（本地 Docker MySQL，宿主机 3307）"],
      ["node tools/run-api.mjs --kill-ports", "启动 API 前清理占用 8080 的旧进程"],
      ["node tools/run-web.mjs", "启动前端 Vite 开发服务器"],
      ["node tools/run-web.mjs --kill-ports", "启动前端前清理占用 5173 的旧进程"],
    ],
  },
  {
    title: "本地依赖",
    commands: [
      ["node tools/dev-up.mjs", "启动 MySQL 和 MinIO 容器"],
      ["node tools/dev-down.mjs", "停止本地依赖容器，不删除 volume 数据"],
      ["node tools/dev-down.mjs --volumes", "停止容器并删除 volume 数据"],
      ["node tools/dev-status.mjs", "查看本地依赖容器状态"],
      ["node tools/dev-logs.mjs", "查看本地依赖容器日志"],
      ["node tools/dev-logs.mjs --follow", "持续跟随本地依赖容器日志"],
    ],
  },
  {
    title: "检查",
    commands: [
      ["node tools/check.mjs", "全量检查：Go 测试、前端检查/构建、OpenSpec 校验"],
      ["node tools/test-server.mjs", "只跑 Go 后端测试"],
      ["node tools/check-web.mjs", "只跑前端 TypeScript 检查和构建"],
    ],
  },
  {
    title: "日常部署",
    commands: [
      ["node tools/publish-acr-images.mjs", "本地：构建并推送 API/Worker/Web 镜像，生成 deploy/releases/staging-current.env"],
      [
        "scp deploy/releases/staging-current.env root@120.26.28.65:/root/repos/memotree/deploy/releases/staging-current.env",
        "本地：把 release env 传到 staging 服务器",
      ],
      ["cd /root/repos/memotree", "服务器：进入项目目录"],
      ["git pull", "服务器：拉取最新代码和部署脚本"],
      ["sh deploy/staging-deploy.sh deploy/releases/staging-current.env", "服务器：按 release env 拉镜像、重启服务并健康检查"],
    ],
  },
  {
    title: "单脚本帮助",
    commands: [
      ["node tools/dev.mjs --help", "查看一键开发脚本参数"],
      ["node tools/run-api.mjs --help", "查看 API 启动脚本参数"],
      ["node tools/run-web.mjs --help", "查看前端启动脚本参数"],
      ["node tools/dev-up.mjs --help", "查看依赖启动脚本说明"],
      ["node tools/dev-down.mjs --help", "查看依赖停止脚本说明"],
      ["node tools/dev-logs.mjs --help", "查看依赖日志脚本说明"],
      ["node tools/check.mjs --help", "查看全量检查脚本说明"],
    ],
  },
];

// 长命令单独换行展示，避免一条 scp 命令把所有分组的对齐宽度撑得很散。
const commandWidth = Math.min(
  42,
  Math.max(...sections.flatMap((section) => section.commands.map(([command]) => command.length))),
);

console.log("MemoTree tools\n");
for (const section of sections) {
  console.log(section.title);
  for (const [command, description] of section.commands) {
    if (command.length > commandWidth) {
      console.log(`  ${command}`);
      console.log(`  ${"".padEnd(commandWidth)}  ${description}`);
    } else {
      console.log(`  ${command.padEnd(commandWidth)}  ${description}`);
    }
  }
  console.log("");
}
