#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
BIN_PATH="$BACKEND_DIR/bin/netdisk-server"
RUN_DIR="$BACKEND_DIR/tmp"
PID_PATH="$RUN_DIR/netdisk-server.pid"
LOG_PATH="$RUN_DIR/netdisk-server.log"
LEGACY_PID_FILES=("$BACKEND_DIR/server.pid")

# ---- Usage ----
#   ./build-run.sh                 build only, binary at backend/bin/netdisk-server
#   ./build-run.sh run [-- args]   build, stop existing backend, then start it
#   ./build-run.sh restart [-- args]
#                                  same as run
#   ./build-run.sh stop            stop running backend from pid file
#   ./build-run.sh status          show backend status from pid file
#   ./build-run.sh help            show this help
#
#   Examples:
#   ./build-run.sh run
#   ./build-run.sh restart -- --config /etc/netdisk.yaml

usage() {
    sed -n 's/^#   //p' "$0"
}

COMMAND=${1:-build}
case "$COMMAND" in
    build | run | restart | stop | status)
        shift || true
        ;;
    help | -h | --help)
        usage
        exit 0
        ;;
    *)
        COMMAND="build"
        ;;
esac
if [ "${1:-}" = "--" ]; then
    shift
fi

read_pid() {
    local file="$1"
    local pid=""
    if [ -f "$file" ]; then
        pid="$(tr -d '[:space:]' < "$file")"
    fi
    if [[ "$pid" =~ ^[0-9]+$ ]]; then
        printf '%s' "$pid"
    fi
}

is_running() {
    local pid="$1"
    [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null
}

process_exe() {
    local pid="$1"
    local exe=""

    if [ -e "/proc/$pid/exe" ]; then
        exe="$(readlink "/proc/$pid/exe" 2>/dev/null || true)"
        exe="${exe% (deleted)}"
    fi

    printf '%s' "$exe"
}

is_backend_process() {
    local pid="$1"
    local exe

    exe="$(process_exe "$pid")"
    if [ -n "$exe" ]; then
        [ "$exe" = "$BIN_PATH" ] && return 0
        return 1
    fi

    ps -p "$pid" -o args= 2>/dev/null | grep -F -- "$BIN_PATH" >/dev/null
}

pid_files() {
    printf '%s\n' "$PID_PATH" "${LEGACY_PID_FILES[@]}"
}

remove_pid_file() {
    local pid_file="$1"

    rm -f "$pid_file"
}

status_backend() {
    local seen=""
    local pid

    while IFS= read -r pid_file; do
        pid="$(read_pid "$pid_file")"
        if [ -z "$pid" ] || [[ "$seen" == *" $pid "* ]]; then
            [ -z "$pid" ] && remove_pid_file "$pid_file"
            continue
        fi
        seen="${seen} ${pid} "

        if is_running "$pid" && is_backend_process "$pid"; then
            echo "==> Backend running (pid $pid)"
            echo "==> PID: $pid_file"
            echo "==> Log: $LOG_PATH"
            return 0
        fi

        remove_pid_file "$pid_file"
    done < <(pid_files)

    echo "==> Backend not running"
    return 1
}

stop_backend() {
    local stopped=0
    local seen=""

    while IFS= read -r pid_file; do
        local pid
        pid="$(read_pid "$pid_file")"
        if [ -z "$pid" ] || [[ "$seen" == *" $pid "* ]]; then
            remove_pid_file "$pid_file"
            continue
        fi
        seen="${seen} ${pid} "

        if is_running "$pid"; then
            if ! is_backend_process "$pid"; then
                echo "==> PID file points to another process, ignoring: $pid_file (pid $pid)"
                remove_pid_file "$pid_file"
                continue
            fi

            echo "==> Stopping backend (pid $pid)"
            kill "$pid" 2>/dev/null || true
            for _ in {1..30}; do
                if ! is_running "$pid"; then
                    break
                fi
                sleep 0.2
            done
            if is_running "$pid"; then
                echo "==> Backend did not stop gracefully, killing (pid $pid)"
                kill -9 "$pid" 2>/dev/null || true
            fi
            stopped=1
        fi
        remove_pid_file "$pid_file"
    done < <(pid_files)

    if [ "$stopped" -eq 0 ]; then
        echo "==> No running backend found"
    fi
}

start_backend() {
    mkdir -p "$BACKEND_DIR/bin" "$RUN_DIR"
    echo "==> Starting backend in background..."
    cd "$BACKEND_DIR"
    nohup "$BIN_PATH" "$@" > "$LOG_PATH" 2>&1 &
    local pid=$!

    sleep 0.5
    if ! is_running "$pid" || ! is_backend_process "$pid"; then
        rm -f "$PID_PATH"
        echo "==> Backend failed to start, see log: $LOG_PATH" >&2
        tail -n 40 "$LOG_PATH" >&2 || true
        exit 1
    fi

    echo "$pid" > "$PID_PATH"
    echo "==> Backend started (pid $pid)"
    echo "==> PID: $PID_PATH"
    echo "==> Log: $LOG_PATH"
}

if [ "$COMMAND" = "stop" ]; then
    stop_backend
    exit 0
fi

if [ "$COMMAND" = "status" ]; then
    status_backend
    exit $?
fi

echo "==> 1/4: Install frontend dependencies"
cd "$ROOT_DIR/frontend"
corepack enable
pnpm install --frozen-lockfile

echo "==> 2/4: Build frontend (static SPA)"
cp svelte.config.js svelte.config.js.bak
cp svelte.config.static.js svelte.config.js
restore_svelte_config() {
    if [ -f svelte.config.js.bak ]; then
        mv svelte.config.js.bak svelte.config.js
    fi
}
trap restore_svelte_config EXIT
pnpm build
restore_svelte_config
trap - EXIT

echo "==> 3/4: Copy frontend build into backend embed path"
rm -rf "$ROOT_DIR/backend/internal/web/build"
mkdir -p "$ROOT_DIR/backend/internal/web/build"
cp -r build/* "$ROOT_DIR/backend/internal/web/build/"

echo "==> 3.5/4: Sanitize for Go embed (rename _ prefixed dirs/files)"
BUILD_DIR="$ROOT_DIR/backend/internal/web/build"
find "$BUILD_DIR" -name '_*' -depth -exec sh -c 'for f; do d=$(dirname "$f"); b=$(basename "$f"); mv "$f" "$d/${b#_}"; done' _ {} +
find "$BUILD_DIR" -type f \( -name '*.html' -o -name '*.js' -o -name '*.css' -o -name '*.json' \) -exec sed -i 's|/_app/|/app/|g' {} +

echo "==> 4/4: Build backend"
cd "$BACKEND_DIR"
go build -trimpath -ldflags="-s -w" -o "$BIN_PATH" ./cmd/server

if [ "$COMMAND" = "run" ] || [ "$COMMAND" = "restart" ]; then
    stop_backend
    start_backend "$@"
fi
