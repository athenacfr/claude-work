---
description: Run development commands (dev servers, watchers, type generators) in the background. Auto-discovers commands on first run. Usage: /dev [stop|restart|status|update|logs]
---

# Dev — Background Development Commands

Run development commands (dev servers, build watchers, type generators) in the background for all subprojects.

## Argument

The argument is: `$ARGUMENTS`

## Config File

Dev commands are persisted at `$CW_TASK_DIR/dev-config.json`:

```json
{
  "subprojects": [
    {
      "path": "frontend",
      "port": 5173,
      "commands": [
        {
          "cmd": "npm run dev",
          "description": "Vite dev server with HMR",
          "type": "long-running"
        },
        {
          "cmd": "npm run generate:types",
          "description": "Generate GraphQL types from schema",
          "type": "one-shot"
        }
      ]
    },
    {
      "path": "backend",
      "venv": ".venv",
      "port": 8000,
      "commands": [
        {
          "cmd": "uvicorn main:app --reload",
          "description": "FastAPI dev server",
          "type": "long-running"
        },
        {
          "cmd": "alembic upgrade head",
          "description": "Run database migrations",
          "type": "one-shot"
        }
      ]
    }
  ]
}
```

**Command types:**
- `one-shot` — runs once and completes (codegen, migrations, builds). Executed first, sequentially.
- `long-running` — runs continuously (dev servers, watchers, file watchers). Launched in parallel as background tasks.

**Optional fields per subproject:**
- `venv` (string) — path to a Python virtual environment relative to the subproject root (e.g. `.venv`). When set, all commands for this subproject are prefixed with `source <venv>/bin/activate &&`.
- `port` (number) — the port this subproject's dev server listens on. Used for port conflict detection and env override updates.

**Top-level optional field:**
- `portBase` (number) — the base port for this project's port range. Each subproject gets ports offset from this base. This prevents port conflicts between different cw projects.

## Port Allocation

Each cw project gets a deterministic port range to avoid conflicts when multiple projects run simultaneously.

**Scheme:**
- During first discovery, assign a `portBase` derived from a hash of the project name, mapped to a range (e.g. 3000-9999). Example: project "myapp" → portBase 4200.
- Each subproject gets the next port in sequence: first subproject uses `portBase`, second uses `portBase + 1`, etc.
- The `port` field on each subproject stores its assigned port.
- Commands that need a port (dev servers) should be launched with the assigned port via flag (e.g. `--port 4200`) or env var.

**Config example with portBase:**
```json
{
  "portBase": 4200,
  "subprojects": [
    { "path": "frontend", "port": 4200, ... },
    { "path": "backend",  "port": 4201, "venv": ".venv", ... }
  ]
}
```

**During discovery**, after identifying subproject commands:
1. Compute a `portBase` from the project name (hash mod 5000 + 3000, giving range 3000-7999)
2. Assign sequential ports to subprojects that have long-running dev servers
3. Modify the discovered commands to use the assigned port (e.g. append `--port <N>` or set `PORT=<N>` prefix)
4. Show the port assignments in the discovery confirmation so the user can adjust

**Port flag conventions by stack:**
- Node.js/Vite: `--port <N>` or `PORT=<N>`
- Python/uvicorn: `--port <N>`
- Python/Django: `0.0.0.0:<N>` as positional arg to runserver
- Go/air/custom: `PORT=<N>` env var prefix
- Cargo: `PORT=<N>` env var prefix

## Process

### /dev (no argument or first run)

#### If config exists — launch

1. Read `$CW_TASK_DIR/dev-config.json`
2. **Port conflict check**: For each subproject with a `port` field, check if that port is already in use:
   ```bash
   lsof -i :<port> -sTCP:LISTEN -t 2>/dev/null
   ```
   If a port is occupied, warn the user and suggest an alternative port. If the user picks a different port, update the command accordingly (e.g. append `--port <new-port>`) and update the config file.
3. **Env override sync**: For **every** subproject with a `port` field, check all env override files at `$CW_PROJECT_DIR/.env.<repo>.override` for variables whose values contain a port number that should match this subproject's configured port. Look for URL-shaped values (e.g. `http://localhost:<port>`) and bare port variables (e.g. `PORT=<port>`). Common variable names: `API_URL`, `BACKEND_URL`, `VITE_API_URL`, `PORT`, `NEXT_PUBLIC_API_URL`, `DATABASE_URL`. If any value references a different port than what's configured, update the override file so all subprojects connect to the right address. This runs on every launch, not just when ports change from conflicts. The env sync to `.env` files happens automatically via cw's file watcher — no manual sync step needed.
5. For each subproject, run one-shot commands first (sequentially, wait for each to complete). **Redirect output to log files**:
   - If the subproject has a `venv` field:
     ```bash
     cd <project-dir>/<subproject-path> && source <venv>/bin/activate && <one-shot-cmd> >> "$CW_TASK_DIR/logs/<subproject>.log" 2>&1
     ```
   - Otherwise:
     ```bash
     cd <project-dir>/<subproject-path> && <one-shot-cmd> >> "$CW_TASK_DIR/logs/<subproject>.log" 2>&1
     ```
   If a one-shot command fails (non-zero exit), read the last 20 lines of its log file to show the error, then ask if the user wants to continue or abort.
6. Then launch all long-running commands in parallel using `run_in_background: true`. **Redirect all output to log files** so it doesn't accumulate in Claude's memory:
   - If the subproject has a `venv` field:
     ```bash
     cd <project-dir>/<subproject-path> && source <venv>/bin/activate && <long-running-cmd> >> "$CW_TASK_DIR/logs/<subproject>.log" 2>&1
     ```
   - Otherwise:
     ```bash
     cd <project-dir>/<subproject-path> && <long-running-cmd> >> "$CW_TASK_DIR/logs/<subproject>.log" 2>&1
     ```
7. Display a summary table:
   ```
   Dev commands running:

   Subproject   Command              Type          Port   Status
   ─────────────────────────────────────────────────────────────
   frontend     npm run dev          long-running  :5173  ✓ background
   backend      uvicorn main:app     long-running  :8000  ✓ background (venv)
   backend      alembic upgrade head one-shot       —     ✓ completed (venv)

   URLs:
     frontend  → http://localhost:5173
     backend   → http://localhost:8000

   Logs: .cw/logs/frontend.log, .cw/logs/backend.log
   Use /dev logs to view output, /dev status to check health.
   ```

8. After displaying the table, show a **URLs section** listing each subproject that has a port with its URL as `http://localhost:<port>`. Only include subprojects with long-running commands that have a port assigned.

#### If NO config exists — discover and confirm

1. List all subdirectories in the project root (these are subprojects)
2. For each subproject, look for:
   - `package.json` → check `scripts` for dev, start, watch, generate, build:watch, codegen, typecheck entries
   - `Makefile` → check for dev, watch, serve, run, generate targets
   - `Cargo.toml` → cargo watch, cargo run
   - `go.mod` → check Makefile or common go run/air/templ patterns
   - `pyproject.toml` / `manage.py` → check for runserver, celery, uvicorn patterns. Also check if `.venv/` or `venv/` exists — if so, set `venv` field in the config so commands are activated properly.
   - `docker-compose.yml` → check for dev services
   - `Procfile` / `Procfile.dev` → dev process definitions
3. For each discovered command, classify as `one-shot` or `long-running`:
   - **long-running**: dev, start, watch, serve, runserver (anything that keeps running)
   - **one-shot**: generate, codegen, build, typecheck, migrate (anything that completes)
4. Present the discovered config to the user using **AskUserQuestion**:
   ```
   Discovered dev commands:

   frontend/ (Node.js) — port 5173
     - npm run dev          → Vite dev server [long-running]
     - npm run generate     → Generate GraphQL types [one-shot]

   backend/ (Python, venv: .venv) — port 8000
     - uvicorn main:app --reload  → FastAPI dev server [long-running]
     - alembic upgrade head       → Run migrations [one-shot]

   Does this look right? You can:
   - Confirm to save and start
   - Add/remove/modify commands
   - Change ports or venv paths
   - Skip a subproject
   ```
5. If user confirms, write the config to `$CW_TASK_DIR/dev-config.json` and launch (go to "If config exists" flow)
6. If user wants changes, adjust and confirm again

### /dev stop

1. Stop all running background dev tasks using the TaskStop tool
2. Confirm: "All dev commands stopped." (Log files persist at `.cw/logs/` — cw cleans them up automatically when the session ends.)

### /dev restart

1. Stop all running background dev tasks using the TaskStop tool
2. Clear log files for a fresh start:
   ```bash
   rm -f "$CW_TASK_DIR/logs/"*.log
   ```
3. Re-launch everything from config — run the full "If config exists" flow (port conflict check, env override sync, one-shot commands, then long-running commands)
4. Display the summary table and confirm: "Dev commands restarted."

### /dev update

Re-discover and merge changes into the existing config without losing manual edits.

1. Read the existing config from `$CW_TASK_DIR/dev-config.json`
2. Run the full discovery process (same as "If NO config exists" above)
3. Diff the discovered config against the existing config and present changes using **AskUserQuestion**:
   ```
   Config update — changes detected:

   NEW subprojects:
     + worker/ (Python, venv: .venv) — port 4202
       - celery -A app worker  → Celery worker [long-running]

   CHANGED subprojects:
     ~ frontend/ — 1 new command found
       + npm run typecheck      → TypeScript check [one-shot]

   REMOVED subprojects (no longer detected):
     - legacy-api/  (still in config — keep or remove?)

   UNCHANGED:
     = backend/ — no changes

   Accept changes? You can:
   - Accept all
   - Accept selectively
   - Edit before saving
   ```
4. Merge accepted changes into the existing config, preserving:
   - Manual edits to commands, descriptions, and types
   - Custom `venv` paths and `port` assignments
   - The existing `portBase` (assign new subprojects the next available port in sequence)
5. Write updated config to `$CW_TASK_DIR/dev-config.json`
6. If dev commands are currently running, ask: "Restart with updated config?"

### /dev status

1. Check if background tasks are still running using TaskOutput with `block: false` (non-blocking check — just get status, don't wait)
2. Display status for each:
   ```
   Dev command status:

   Subproject   Command          Port   Status
   ──────────────────────────────────────────────
   frontend     npm run dev      :5173  running
   backend      uvicorn main:app :8000  exited (error)
   ```
3. For any failed or errored tasks, show the **last 10 lines** of the log file (not full output):
   ```bash
   tail -n 10 "$CW_TASK_DIR/logs/<subproject>.log"
   ```
4. Ask if the user wants to restart failed commands
5. Mention: "Use `/dev logs <subproject>` for more output."

### /dev logs [subproject] [lines]

Read dev server logs in controlled chunks to minimize token usage.

1. If no subproject specified, show the **last 50 lines** of each subproject's log file
2. If subproject specified, show the **last 50 lines** of that subproject's log
3. If lines specified, override the default (e.g. `/dev logs backend 200`)
4. Read logs using:
   ```bash
   tail -n <lines> "$CW_TASK_DIR/logs/<subproject>.log"
   ```
Note: Log file size is managed automatically by cw (truncated to last 5000 lines if exceeding 10MB). No need to handle this in the LLM.

## Error Handling

- When a background task completes unexpectedly (crash), you'll get an automatic notification. Read the **last 20 lines** of the log file to surface the error — do NOT use TaskOutput for the full output. Ask if the user wants to restart it.
- If a one-shot command fails during startup, show the error and ask whether to continue with remaining commands or abort.
- If the config file is malformed, show the error and offer to re-discover.

## Important

- Always `cd` to the subproject directory before running commands — never run from project root
- Use absolute paths when constructing the `cd` path: `$CW_PROJECT_DIR/<subproject-path>`
- One-shot commands run sequentially and must complete before long-running commands start
- Long-running commands all run in parallel as background tasks
- The config file is the source of truth — always read it before launching
- If the user modifies the config manually, respect those changes
- When discovering commands, read actual file contents (package.json scripts, Makefile targets) — don't guess
- **Python venv**: During discovery, check for `.venv/` or `venv/` directories. If found, set the `venv` field in config. Always activate the venv before running any Python subproject command — without it, commands will use the system Python and fail to find project dependencies.
- **Port awareness**: During discovery, infer default ports from config files and command flags (e.g. `--port 8000`, Vite's default 5173, Django's default 8000). Store in the `port` field. Before launching, check for port conflicts — if multiple cw sessions or external processes occupy a port, choose an alternative and update env override files so cross-service references stay correct.
- **Env override sync**: When a port changes, scan `$CW_PROJECT_DIR/.env.<repo>.override` files for URL or port variables referencing the old port and update them. This ensures that e.g. a frontend's `VITE_API_URL` points to the backend's actual running port.
- **Log management**: All dev process output goes to `$CW_TASK_DIR/logs/<subproject>.log`. NEVER use TaskOutput to read full process output — always read log files with `tail` to control token usage. Log cleanup (deletion on session end, truncation of oversized files) is handled automatically by cw — do NOT run cleanup commands yourself.
