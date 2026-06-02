#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker 未安装。请先安装 Docker Engine 和 Docker Compose plugin 后再运行本脚本。"
  echo "Ubuntu/Debian 可参考：curl -fsSL https://get.docker.com | sudo sh"
  exit 1
fi

if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE=(docker-compose)
else
  echo "Docker Compose 未安装。请安装 Docker Compose plugin 后再运行本脚本。"
  exit 1
fi

if [ ! -f .env ] && [ -f .env.example ]; then
  cp .env.example .env
fi

"${COMPOSE[@]}" up -d --build
"${COMPOSE[@]}" ps

echo
echo "启动完成：访问 http://服务器公网IP:${APP_PORT:-5290}"
echo "查看日志：${COMPOSE[*]} logs -f app"
