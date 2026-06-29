#!/bin/sh
set -e

# Fix log directory permissions when a bind-mounted host directory
# (created by Docker as root) blocks the application user.
mkdir -p /app/logs
chown netdisk:netdisk /app/logs

exec su-exec netdisk:netdisk netdisk-server "$@"
