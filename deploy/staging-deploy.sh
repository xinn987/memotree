#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

STAGING_ENV=${STAGING_ENV:-deploy/.env.staging}
RELEASE_ENV=${1:-deploy/releases/staging-current.env}
COMPOSE_FILE=${COMPOSE_FILE:-deploy/docker-compose.staging.yml}

usage() {
  cat <<'EOF'
Usage:
  sh deploy/staging-deploy.sh [release-env-file]

Example:
  sh deploy/staging-deploy.sh deploy/releases/staging-current.env

The release env file must define API_IMAGE, WORKER_IMAGE, and WEB_IMAGE.
Long-lived secrets stay in deploy/.env.staging.
EOF
}

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

require_file() {
  [ -f "$1" ] || fail "missing required file: $1"
}

load_release_env() {
  # 只加载发布版本相关字段，避免误把长期密钥从版本单带入部署进程。
  while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in
      ""|\#*) continue ;;
    esac

    key=${line%%=*}
    value=${line#*=}
    case "$key" in
      API_IMAGE|WORKER_IMAGE|WEB_IMAGE|RELEASE_COMMIT|RELEASE_IMAGE_TAG|RELEASE_CREATED_AT)
        [ "$value" != "" ] || fail "$key is empty in $RELEASE_ENV"
        export "$key=$value"
        ;;
      *)
        fail "unsupported key in release env: $key"
        ;;
    esac
  done < "$RELEASE_ENV"
}

compose() {
  docker compose --env-file "$STAGING_ENV" -f "$COMPOSE_FILE" "$@"
}

check_http() {
  name=$1
  url=$2
  if command -v curl >/dev/null 2>&1; then
    curl -fsS "$url" >/dev/null || fail "$name health check failed: $url"
    echo "$name health check OK: $url"
    return
  fi
  echo "curl not found; skipping $name health check: $url"
}

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
  usage
  exit 0
fi

require_file "$STAGING_ENV"
require_file "$RELEASE_ENV"
require_file "$COMPOSE_FILE"
load_release_env

: "${API_IMAGE:?API_IMAGE is required}"
: "${WORKER_IMAGE:?WORKER_IMAGE is required}"
: "${WEB_IMAGE:?WEB_IMAGE is required}"

echo "Deploying staging release:"
echo "  API_IMAGE=$API_IMAGE"
echo "  WORKER_IMAGE=$WORKER_IMAGE"
echo "  WEB_IMAGE=$WEB_IMAGE"
echo "  RELEASE_COMMIT=${RELEASE_COMMIT:-unknown}"
echo "  RELEASE_CREATED_AT=${RELEASE_CREATED_AT:-unknown}"

echo
echo "Validating Docker Compose config..."
compose config >/dev/null

echo
echo "Pulling business images..."
compose pull api worker web init-storage

echo
echo "Restarting business services..."
compose up -d --no-build api worker web

echo
echo "Current service status:"
compose ps

echo
echo "Running health checks..."
check_http "web" "http://127.0.0.1/healthz"
check_http "api" "http://127.0.0.1/api/healthz"

echo
echo "Staging deploy complete."
echo "Use sh deploy/staging-logs.sh --tail=120 to inspect recent logs."
