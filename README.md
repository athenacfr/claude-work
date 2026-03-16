# iara

**IARA — Inspira Agent Runtime Architecture.** Stop copy-pasting context into Claude Code. iara sets up the right project, the right repos, the right mode, and the right session — then launches Claude with everything it needs to be useful from the first message.

If you work across multiple repos, juggle different tasks in parallel, or find yourself re-explaining your codebase every time you start a new Claude session, iara fixes that.

## The problem

Claude Code is powerful, but it starts every session blank. You `cd` into a directory, launch it, and spend the first few messages explaining what you're working on, which repos matter, and how things fit together. If you're working on multiple things, you're constantly context-switching — and so is Claude.

## What iara does

iara sits between you and Claude Code. It manages **projects** (groups of repos), **tasks** (isolated branches with their own session history), and **modes** (behavioral presets like code, research, or review). You pick what you're working on from a fast TUI, and Claude launches already knowing the full picture.

```
you → iara (pick project, task, mode) → Claude Code (with full context)
```

No setup prompts. No re-explaining. Just start working.

## Quick start

Requires Go 1.24+ and [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

```sh
git clone https://github.com/ahtwr/iara.git
cd iara
make install
```

Then run:

```sh
iara
```

## Why iara

### Work across multiple repos

Real projects span multiple repositories — backend, frontend, shared libraries, infra. iara groups them into a single project so Claude sees everything, not just the directory you happened to `cd` into.

### Parallel tasks without conflicts

Every task gets its own **git worktree branch**. Work on a feature in one task, a bugfix in another — each with isolated branches and their own session history. No stashing, no branch juggling. When a task is done, its worktree is cleaned up automatically.

### Sessions that remember

Start a Claude session, close it, come back later — iara tracks your session history per task. Resume exactly where you left off, or start fresh. Your context carries over.

### Modes set the tone

Instead of telling Claude "act as a code reviewer" every time, pick the `review` mode once. Modes shape Claude's behavior for the entire session:

| Mode       | What Claude does                      |
| ---------- | ------------------------------------- |
| `code`     | Writes features, fixes bugs (default) |
| `research` | Explores the codebase, read-only      |
| `review`   | Reviews code changes                  |
| `none`     | No preset — do whatever you want      |

Switch modes mid-session with `/mode research` without restarting.

### Dev servers, managed

Type `/dev` inside Claude and it auto-discovers dev commands in your repos (`package.json` scripts, `Makefile` targets, `Cargo.toml`, etc.), assigns deterministic ports, and runs them in the background. No more "what port was that on?" — it's the same every time.

### Go autonomous when you want

`/yolo` lets Claude plan and execute a full objective autonomously — map the codebase, break the work into tasks, and execute without stopping to ask at every step. It's opt-in, explicit, and useful for well-scoped work you trust Claude to handle.

## Commands inside Claude

Once you're in a Claude session, these slash commands extend what you can do without leaving:

| Command                         | What it does                          |
| ------------------------------- | ------------------------------------- |
| `/mode [name]`                  | Show or switch behavioral mode        |
| `/permissions [bypass\|normal]` | Toggle Claude's permission prompts    |
| `/dev`                          | Start dev servers and watchers        |
| `/dev stop`                     | Stop all dev servers                  |
| `/yolo [objective]`             | Autonomous plan-and-execute           |
| `/new-task`                     | Create a new task with its own branch |
| `/finish-task`                  | Complete task, clean up worktrees     |
| `/iara:compact-and-continue`      | Compact context and keep working      |
| `/iara:new-session`               | Start a fresh session                 |
| `/iara:reload`                    | Reload config and commands            |
| `/iara:help`                      | Show all available commands           |

## How it works

```
~/iara/projects/my-app/
├── backend/              ← git repo
├── frontend/             ← git repo
├── .worktrees/
│   └── fix-auth/         ← isolated worktree for a task
│       ├── backend/
│       └── frontend/
├── .iara/
│   ├── metadata.json     ← project title, description, instructions
│   └── tasks/
│       └── <id>/
│           ├── task.json
│           └── sessions/ ← session history for this task
└── CLAUDE.md             ← auto-generated project context
```

Projects live under `~/iara/projects/`. Each project contains repos as subdirectories, task worktrees for isolated branches, and metadata that iara injects into Claude's system prompt on every launch.

## Environment variables

| Variable          | Description                             |
| ----------------- | --------------------------------------- |
| `IARA_PROJECTS_DIR` | Override the default projects directory |

## Uninstall

```sh
iara uninstall
```

Removes the binary and all project data.
