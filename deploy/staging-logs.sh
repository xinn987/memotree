#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
cd "$REPO_ROOT"

STAGING_ENV=${STAGING_ENV:-deploy/.env.staging}
COMPOSE_FILE=${COMPOSE_FILE:-deploy/docker-compose.staging.yml}
TAIL=120
FOLLOW=0
SERVICES="api worker web mysql"
SERVICES_SET=0

usage() {
  cat <<'EOF'
Usage:
  sh deploy/staging-logs.sh [--follow] [--tail=N] [service...]

Examples:
  sh deploy/staging-logs.sh
  sh deploy/staging-logs.sh --tail=200 api worker
  sh deploy/staging-logs.sh --follow
EOF
}

for arg in "$@"; do
  case "$arg" in
    --help|-h)
      usage
      exit 0
      ;;
    --follow|-f)
      FOLLOW=1
      ;;
    --tail=*)
      TAIL=${arg#--tail=}
      ;;
    *)
      if [ "$SERVICES_SET" -eq 0 ]; then
        SERVICES=""
        SERVICES_SET=1
      fi
      SERVICES="$SERVICES $arg"
      ;;
  esac
done

[ -f "$STAGING_ENV" ] || {
  echo "ERROR: missing required file: $STAGING_ENV" >&2
  exit 1
}

args="logs --tail=$TAIL"
if [ "$FOLLOW" -eq 1 ]; then
  args="$args -f"
fi

# shellcheck disable=SC2086
docker compose --env-file "$STAGING_ENV" -f "$COMPOSE_FILE" $args $SERVICES
