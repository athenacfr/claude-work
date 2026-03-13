#!/usr/bin/env bash
# description: Reload the cw session to pick up new commands, rules, and config changes.
echo "Reloading session..."
cw internal reload
