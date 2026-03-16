#!/usr/bin/env bash
# Stop hook for yolo mode. Prevents Claude from exiting during autonomous execution.
# Only active when IARA_YOLO_ACTIVE=1 — does nothing in normal sessions.

[ "$IARA_YOLO_ACTIVE" != "1" ] && exit 0

# Consume stdin (hook input JSON)
cat > /dev/null

PLAN="${IARA_YOLO_PLAN:-}"
[ -z "$PLAN" ] && exit 0
[ ! -f "$PLAN" ] && exit 0

# Block exit and re-feed the plan
jq -n --arg plan "$PLAN" '{
  decision: "block",
  reason: ("Continue working on the plan at " + $plan + ". Read it, check off completed tasks, work on remaining [ ] items.")
}'
