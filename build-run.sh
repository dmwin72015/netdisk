#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ---- Usage ----
#   ./build-run.sh          build only, binary at backend/bin/netdisk-server
#   ./build-run.sh run      build then start backend
#   ./build-run.sh run --   build then start backend with extra flags (e.g. --config /etc/netdisk.yaml)

RUN=${1:-}
if [ "$RUN" = "run" ]; then
    shift
fi

echo "==> 1/4: Install frontend dependencies"
cd "$ROOT_DIR/frontend"
corepack enable
pnpm install --frozen-lockfile

echo "==> 2/4: Build frontend (static SPA)"
cp svelte.config.js svelte.config.js.bak
cp svelte.config.static.js svelte.config.js
pnpm build
mv svelte.config.js.bak svelte.config.js

echo "==> 3/4: Copy frontend build into backend embed path"
rm -rf "$ROOT_DIR/backend/internal/web/build"
mkdir -p "$ROOT_DIR/backend/internal/web/build"
cp -r build/* "$ROOT_DIR/backend/internal/web/build/"

echo "==> 3.5/4: Sanitize for Go embed (rename _ prefixed dirs/files)"
BUILD_DIR="$ROOT_DIR/backend/internal/web/build"
find "$BUILD_DIR" -name '_*' -depth -exec sh -c 'for f; do d=$(dirname "$f"); b=$(basename "$f"); mv "$f" "$d/${b#_}"; done' _ {} +
find "$BUILD_DIR" -type f \( -name '*.html' -o -name '*.js' -o -name '*.css' -o -name '*.json' \) -exec sed -i 's|/_app/|/app/|g' {} +

echo "==> 4/4: Build backend"
cd "$ROOT_DIR/backend"
go build -trimpath -ldflags="-s -w" -o bin/netdisk-server ./cmd/server

if [ "$RUN" = "run" ]; then
    echo "==> Starting backend..."
    exec ./bin/netdisk-server "$@"
fi
