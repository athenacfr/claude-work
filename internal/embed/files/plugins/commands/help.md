---
description: Show all iara commands and modes available inside Claude.
---

# CW Help

Show the user all available iara commands and current session info.

## Process

Display the following:

```
CW Commands
═══════════

/mode                   Show current mode
/mode <name>            Switch mode (code, research, review, plan, tdd, debug, free)

/permissions            Show current permissions mode
/permissions <value>    Switch permissions (bypass, normal)

/iara:compact-and-continue  Compact context and continue where you left off
/iara-help                Show this help

Current Session
═══════════════
Project:   $IARA_PROJECT (from env var, or "unknown")
Mode:      $IARA_MODE (from env var, or "code")
Directory: (run pwd)
Branch:    (run git branch --show-current)
```

Read the `IARA_PROJECT` and `IARA_MODE` environment variables using Bash to populate the current session info. If they're not set, show "unknown" and "code" respectively.
