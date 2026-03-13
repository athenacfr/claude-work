#!/usr/bin/env bash
# Auto-compact hook: counts PostToolUse calls and triggers compact
# when the estimated context usage exceeds CW_AUTO_COMPACT_LIMIT%.

LIMIT="${CW_AUTO_COMPACT_LIMIT:-0}"
[ "$LIMIT" -eq 0 ] 2>/dev/null && exit 0

# Use a temp file scoped to the cw process to track call count
COUNTER_FILE="/tmp/cw-compact-counter-${CW_PID:-$$}"

# Read and increment counter
COUNT=0
[ -f "$COUNTER_FILE" ] && COUNT=$(cat "$COUNTER_FILE" 2>/dev/null || echo 0)
COUNT=$((COUNT + 1))
echo "$COUNT" > "$COUNTER_FILE"

# Map limit % to approximate tool call threshold
# Conservative estimates: ~100 tool calls ≈ full context
THRESHOLD=$(( LIMIT ))

if [ "$COUNT" -ge "$THRESHOLD" ]; then
  # Reset counter and trigger auto-compact
  rm -f "$COUNTER_FILE"
  cw internal auto-compact &
fi

exit 0
