---
description: Plan and execute work autonomously. Usage: /yolo [objective]
---

# Yolo Mode — Autonomous Planning & Execution

Plan a set of tasks and then execute them autonomously without human intervention.

## Argument

The argument is: `$ARGUMENTS`

## Process

### Step 1: Determine Objective

**If argument provided** (e.g., `/yolo implement auth token refresh`):
- Use the argument as the objective

**If no argument** (`/yolo`):
1. Check for an existing plan file by running: `ls $CW_PROJECT_DIR/.cw/yolo/plan-*.md 2>/dev/null` using the Bash tool
   - **If plan exists**: Use AskUserQuestion to ask: "Active yolo plan found. Resume execution or re-plan?"
     - Resume → skip to Step 5 (start executing)
     - Re-plan → continue to Step 2
2. If no plan exists, analyze the recent conversation context
   - **If relevant context exists** → propose a plan based on what's been discussed
   - **If no context** → ask the user what they want to accomplish

### Step 2: Explore & Plan

1. Explore the codebase to understand what's needed for the objective
2. Ask clarifying questions if the objective is ambiguous (keep it brief — 1-2 questions max)
3. Build a plan with atomic, ordered tasks using `[ ]` checkboxes

### Step 3: Write Plan File

Create the directory and write the plan:

```bash
mkdir -p "$CW_PROJECT_DIR/.cw/yolo"
```

Then write the plan file to `$CW_PROJECT_DIR/.cw/yolo/plan-$CW_SESSION_ID.md` using the Write tool.

Plan format:

```markdown
# Yolo Plan

## Objective
Brief description of what we're building.

## Tasks
- [ ] First task to do
- [ ] Second task
  - [ ] Subtask if needed
- [ ] Third task

## Notes
Any context, decisions, or observations.
```

### Step 4: Confirm

Use AskUserQuestion to show:
- The objective
- The task list (summarized if long)
- Estimated scope (e.g., "~8 tasks, touching 5 files")

Ask: **"Ready to start yolo?"**

- **Yes** → run `cw internal yolo-start` using the Bash tool, then continue to Step 5
- **No** → refine the plan based on feedback, update the plan file, and ask again

### Step 5: Execute

Work through all pending tasks in the plan file autonomously.

**Never ask questions.** Do not use AskUserQuestion. Make decisions yourself.
**Never stop to wait for input.** Keep working until all tasks are done.

1. Read the plan file
2. Find the first unchecked `[ ]` task
3. Implement it
4. Verify it works (run tests, build, lint as appropriate)
5. Check it off `[x]` in the plan file
6. Git commit if you've made meaningful progress
7. Move to the next `[ ]` task
8. Repeat until all tasks are `[x]`

You can add, modify, reorder, or remove tasks as you learn things. Add notes to the Notes section.

**When stuck:** Try a different approach. If you've tried 3 times, skip the task with a note and move on.

**Agents:** Use the Agent tool for parallel or focused work — researcher for exploration, tester for tests, implementer for independent subtasks.

**When ALL tasks are done:** Run `cw internal yolo-stop` using the Bash tool. Do NOT call yolo-stop until every task is done.

## Important

- Tasks should be atomic — one clear deliverable per task
- Order tasks by dependency — things that must be done first come first
- Include verification tasks (run tests, build, lint) where appropriate
- The plan file is a living document — modify it during execution
