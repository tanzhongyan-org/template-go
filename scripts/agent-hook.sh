#!/usr/bin/env bash
set -euo pipefail

EVENT="${1:-}"

case "$EVENT" in
  file-edit)
    FILE="${2:-}"
    if [ -n "$FILE" ] && [ -f "$FILE" ]; then
      case "$FILE" in
        *.go)
          gofmt -w "$FILE" 2>/dev/null || true
          ;;
      esac
    fi
    ;;
  finish-turn)
    if [ "${AGENT_HOOK_CHECKED:-0}" = "1" ]; then
      exit 0
    fi
    export AGENT_HOOK_CHECKED=1
    make check
    ;;
  *)
    exit 0
    ;;
esac