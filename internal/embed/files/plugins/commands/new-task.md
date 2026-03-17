---
description: Create a new task: map codebase, understand intent, create worktree branches, and set up task context.
---

# New Task

Set up a new task by understanding what the user wants to work on, creating git worktrees with branches, and saving task context.

## Environment

- `IARA_PROJECT_DIR` — the project root directory
- `IARA_TASK_ID` — the task ID (set after save-task)

## Process

### Step 1: Map the codebase

Explore all subprojects in the project directory autonomously — do NOT ask the user about tech stack or structure:

- List top-level files and directories in each subproject
- Read package.json, go.mod, Cargo.toml, pyproject.toml, requirements.txt, Makefile, docker-compose.yml, or whatever dependency/config files exist
- Scan a few key source files to understand patterns (naming, formatting, test structure)
- Check for existing linter configs (.eslintrc, .prettierrc, .golangci.yml, etc.)
- Check for CI configs (.github/workflows/, .gitlab-ci.yml, etc.)
- Look at git log for commit message style

### Step 2: Ask what the user wants to work on

Ask: **"What are you working on?"**

This is the only question you ask unprompted. Wait for their answer.

### Step 3: Confirm the intention description

Based on their answer and your codebase understanding, write a clear **description** of the work intention. This should describe what will be done, not summarize what was said.

Present the description and use the **AskUserQuestion** tool to confirm it's correct. If the user wants changes, adjust and confirm again.

### Step 4: Decide and confirm branch names

For each subproject, look at existing branches to detect the naming pattern:

```bash
git -C <subproject> branch -a --format='%(refname:short)' | head -30
```

Common patterns: `feat/...`, `feature/...`, `fix/...`, `chore/...`, flat names like `add-auth`. Match whatever the repo already uses. If no clear pattern, use `feat/<slug>`.

Present the branch name and use **AskUserQuestion** to confirm. All repos will use the same branch name.

### Step 5: Save task and create worktrees

Build a JSON object and run this command using the Bash tool:

```bash
iara internal save-task '{"name":"<slug>","description":"<description>","branch":"<branch-name>"}'
```

This creates the task, sets up git worktrees for each repo, creates the task CLAUDE.md, and symlinks project rules.

### Step 6: Save project metadata (first task only)

Check if project metadata already exists:

```bash
cat "$IARA_PROJECT_DIR/.iara/metadata.json" 2>/dev/null
```

If the file doesn't exist or is empty, this is the first task. Build a JSON object with technical context and save it:

```bash
iara internal save-metadata '{"title":"<title>","description":"<description>","instructions":"<technical-context>"}'
```

The instructions field should contain: tech stack, build/test commands, conventions, coding patterns — everything you learned from mapping the codebase. Write it as direct instructions to Claude.

If metadata already exists, skip this step.

### Step 7: Finish

Check how this command was invoked:

```bash
echo $IARA_AUTO_SETUP
```

- If `1`: Say "All set!" then run `iara internal exit-to-tui`
- Otherwise: Say "All set! Reloading session..." then run `iara internal reload`

## Important

- Derive everything technical from the subprojects. Only ask the user what they're working on.
- Do NOT mention internal files (task.json, metadata.json) to the user.
- Do NOT create any files directly — use the CLI commands.
- All repos share the same branch name for a task.
- Worktrees are created by the save-task command — do NOT run git worktree commands yourself.
