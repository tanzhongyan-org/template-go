#!/bin/sh
# Agent hook entry point for Claude Code and Gemini CLI.
#   post-edit  format the edited Go file
#   stop       run `make verify` when the working tree has changes
# Set AGENT_HOOK_SKIP=1 to skip the stop-mode check.
# Keep stdout empty: Gemini parses stdout as JSON.
set -e

MODE="${1:-post-edit}"
INPUT="$(cat || true)"

if ! command -v make >/dev/null 2>&1; then
  echo "agent-hook: make is not on PATH, so checks were skipped. Install GNU make (Windows: winget install ezwinports.make; macOS: xcode-select --install) and restart the agent." >&2
  exit 0
fi

if [ "$MODE" = "post-edit" ]; then
  FILE_PATH="$2"
  if [ -z "$FILE_PATH" ] && [ -n "$INPUT" ]; then
    FILE_PATH="$(echo "$INPUT" | grep -o '"file_path"[^,]*' | cut -d'"' -f4 || true)"
  fi
  case "$FILE_PATH" in
    *.go)
      if [ -f "$FILE_PATH" ]; then
        gofmt -w "$FILE_PATH" >/dev/null 2>&1 || true
      fi
      ;;
  esac
  exit 0
fi

if [ "$MODE" = "stop" ]; then
  if [ "${AGENT_HOOK_SKIP:-}" = "1" ]; then
    exit 0
  fi
  if echo "$INPUT" | grep -Eq '"stop_hook_active"[[:space:]]*:[[:space:]]*true'; then
    exit 0
  fi
  if [ -z "$(git status --porcelain 2>/dev/null)" ]; then
    exit 0
  fi

  LOG="$(mktemp)"
  trap 'rm -f "$LOG"' EXIT
  if make verify >"$LOG" 2>&1; then
    exit 0
  fi
  {
    echo "make verify failed. Fix these errors before finishing:"
    tail -n 40 "$LOG"
  } >&2
  exit 2
fi
