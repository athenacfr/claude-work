#!/usr/bin/env bash
# description: Switch permissions mode. Usage: /permissions [bypass|normal]
[ -z "$1" ] && exit 1  # no args — defer to LLM to show permissions info
echo "Switching to $1 permissions..."
cw internal permissions-switch "$1"
