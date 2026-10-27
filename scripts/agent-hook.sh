#!/bin/sh
# Agent hook entry point for Claude Code and Gemini CLI.
#   post-edit  format the edited Go file
#   stop       run `make verify` when the working tree has changes
set -e

MODE="${1:-post-edit}"
INPUT="$(cat || true)"

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
  tail -n 40 "$LOG" >&2
  echo "make verify failed. Fix the errors before finishing the turn." >&2
  exit 2
fi
