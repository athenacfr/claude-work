#!/usr/bin/env bash
# description: Switch or show the current behavioral mode. Usage: /mode [code|research|review|plan|tdd|debug|free]
[ -z "$1" ] && exit 1  # no args — defer to LLM to show mode info
echo "Switching to $1 mode..."
cw internal mode-switch "$1"
