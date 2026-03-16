---
description: Complete the current task: verify clean state, remove worktrees, and return to task selection.
---

# Finish Task

Complete the current task by cleaning up worktrees and marking it as done.

## Environment

- `CW_PROJECT_DIR` — the project root directory
- `CW_TASK_ID` — the current task ID
- `CW_TASK_NAME` — the current task name

## Process

### Step 1: Check for uncommitted changes

Check all repos in the worktree for dirty state:

```bash
for dir in "$CW_PROJECT_DIR"/.worktrees/"$CW_TASK_NAME"/*/; do
  if [ -d "$dir/.git" ] || [ -f "$dir/.git" ]; then
    echo "=== $(basename "$dir") ==="
    git -C "$dir" status --porcelain
  fi
done
```

If any repo has uncommitted changes, warn the user and ask if they want to:
1. Commit the changes first
2. Discard changes and proceed
3. Cancel

### Step 2: Confirm

Use **AskUserQuestion**: "Ready to finish task '<task-name>'? The branch will be preserved but the worktree will be removed."

### Step 3: Finish the task

```bash
cw internal finish-task
```

This removes the worktrees and marks the task as completed.

### Step 4: Return to task selection

```bash
cw internal reload
```

## Important

- Always check for uncommitted changes before finishing.
- The git branch is preserved in the original repos — only the worktree working copy is removed.
- The user can push/merge the branch via normal git/gh workflow before or after finishing.
