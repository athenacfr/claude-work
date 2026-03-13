#!/usr/bin/env bash
# Intercepts /cw:* slash commands at prompt submission.
#
# Checks for a matching .sh script in the scripts dir.
# Script exit 0 = handled (block LLM), exit 1 = defer to LLM.

INPUT=$(cat)
PROMPT=$(echo "$INPUT" | jq -r '.prompt // empty')

[ -z "$PROMPT" ] && exit 0

# Match /cw:<command> [args] pattern
if [[ ! "$PROMPT" =~ ^/cw:([a-z-]+)(\ (.+))?$ ]]; then
  exit 0
fi

CMD="${BASH_REMATCH[1]}"
ARGS="${BASH_REMATCH[3]}"

# Check for a direct .sh script
SCRIPTS_DIR="${CW_DATA_DIR:-$HOME/.local/share/cw}/scripts"
SCRIPT="$SCRIPTS_DIR/$CMD.sh"

if [ -x "$SCRIPT" ]; then
  "$SCRIPT" "$ARGS" >&2
  RESULT=$?
  [ "$RESULT" -eq 0 ] && exit 0  # handled
fi

# No script or script deferred — let LLM handle
exit 0
