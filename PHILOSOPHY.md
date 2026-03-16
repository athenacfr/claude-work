# Philosophy

## Claude Code is the IDE

Traditional IDEs are built around human editing — syntax highlighting, autocompletion, file trees. But when an AI agent writes most of the code, the bottleneck shifts from typing to context. The editor doesn't matter as much as what the agent knows when it starts working.

iara treats Claude Code as the primary development environment. The TUI exists only to set up context — projects, repos, tasks, modes — and then get out of the way. Once Claude launches, it has everything it needs: the project structure, the right behavioral mode, isolated worktrees for the current task, and session history to resume from.

## Projects, not files

Most development tools operate at the file or repo level. But real work spans multiple repositories — a backend, a frontend, shared libraries, infrastructure. iara organizes work into projects that group related repos together, so Claude sees the full picture.

## Tasks as isolation boundaries

Every meaningful unit of work gets its own task. A task creates a git worktree branch, giving you an isolated copy of the codebase. Sessions within a task share history, so Claude can pick up where it left off. When the task is done, the worktree is cleaned up. This keeps the main branch clean and lets multiple tasks run in parallel without interference.

## Modes shape behavior

Rather than repeatedly instructing Claude how to behave, modes encode behavioral presets. `research` mode makes Claude read-only and exploratory. `review` mode focuses on code review. `code` mode (the default) is for building. Modes are applied once at launch and persist through the session, reducing prompt overhead.

## Convention over configuration

iara makes decisions so you don't have to:

- Projects go in `~/iara/projects/`
- Dev servers get deterministic ports based on project name
- Environment files layer automatically (global + override)
- First launch triggers auto-setup to map the codebase
- Repos are pulled in the background when you select a project

When defaults work, there's nothing to configure.

## The TUI is a launcher, not a workspace

The TUI is intentionally minimal. It handles project selection, task management, and mode picking — then exits. There's no split panes, no embedded editor, no file browser. The goal is to spend as little time in the TUI as possible and as much time as possible in a Claude session with the right context.

## Slash commands extend the session

Once inside Claude, slash commands (`/dev`, `/yolo`, `/new-task`, `/mode`) let you reshape the session without leaving it. These commands call back into the iara binary to modify state — switching modes, managing dev servers, creating tasks — keeping the feedback loop tight.

## Autonomous when you want it

`/yolo` mode lets Claude plan and execute autonomously: map the codebase, build a task list, and work through it without asking for permission at every step. This is opt-in and explicit. The default mode is collaborative — Claude proposes, you approve.
