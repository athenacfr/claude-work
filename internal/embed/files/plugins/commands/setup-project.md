---
description: Map the codebase and save project metadata for a new cw project.
---

# Setup Project

Set up project metadata by mapping the codebase. This runs automatically on the first launch of a new project.

## Process

### Step 1: Map the codebase

Explore all subprojects in the project directory autonomously — do NOT ask the user about tech stack or structure:

- List top-level files and directories in each subproject
- Read package.json, go.mod, Cargo.toml, pyproject.toml, requirements.txt, Makefile, docker-compose.yml, or whatever dependency/config files exist
- Scan a few key source files to understand patterns (naming, formatting, test structure)
- Check for existing linter configs (.eslintrc, .prettierrc, .golangci.yml, etc.)
- Check for CI configs (.github/workflows/, .gitlab-ci.yml, etc.)
- Look at git log for commit message style

### Step 2: Ask what the project is about

Ask: **"What is this project?"**

Wait for their answer.

### Step 3: Save metadata

Build a JSON object and run:

```bash
cw internal save-metadata '{"title":"<title>","description":"<description>","instructions":"<technical-context>"}'
```

Fields:
- **title**: Short project title
- **description**: What the project is about
- **instructions**: Technical context — structure, tech stack, conventions, build/test commands, coding patterns. Write as direct instructions to Claude.

### Step 4: Finish

Check:

```bash
echo $CW_AUTO_SETUP
```

- If `1`: Say "Project set up! Starting fresh session..." then run `cw internal new-session`
- Otherwise: Say "Project set up! Reloading..." then run `cw internal reload`

## Important

- Derive everything technical from the subprojects. Only ask what the project is about.
- Do NOT mention internal files to the user.
- Do NOT create any files directly — use the CLI command.
- Keep instructions concise but complete.
