#!/bin/sh
set -e

RESTART_INTERVAL_SECONDS="${RESTART_INTERVAL_SECONDS:-86400}"

if [ -n "${RESTART_INTERVAL_SECONDS}" ]; then
  echo "Starting backend (restart interval: ${RESTART_INTERVAL_SECONDS}s)"
  exec timeout "${RESTART_INTERVAL_SECONDS}" ./backend
fi

echo "Starting backend (no restart interval)"
exec ./backend
