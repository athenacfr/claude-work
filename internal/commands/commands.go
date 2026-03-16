package commands

func init() {
	Register(Command{
		Name:        "compact-and-continue",
		Description: "Compact the current session context and continue where you left off.",
		CLICommand:  "compact-and-continue",
	})

	Register(Command{
		Name:        "new-session",
		Description: "Close the current session and start a fresh one with updated configuration.",
		CLICommand:  "new-session",
		Internal:    true,
	})

	Register(Command{
		Name:        "reload",
		Description: "Reload the cw session to pick up new commands, rules, and config changes.",
		CLICommand:  "reload",
		Internal:    true,
	})

	Register(Command{
		Name:        "switch-mode",
		Description: "Switch Claude's behavioral mode mid-session. The session will reload with the new mode's system prompt.",
		CLICommand:  "mode-switch",
		Internal:    true,
		Params: map[string]ParamDef{
			"mode": {
				Type:        "string",
				Description: "The mode to switch to.",
				Enum:        []string{"code", "research", "review", "plan", "tdd", "debug", "free"},
				Required:    true,
			},
		},
	})

	Register(Command{
		Name:        "switch-permissions",
		Description: "Switch Claude's permission mode mid-session. The session will reload with the new permission setting.",
		CLICommand:  "permissions-switch",
		Internal:    true,
		Params: map[string]ParamDef{
			"value": {
				Type:        "string",
				Description: "The permissions mode.",
				Enum:        []string{"bypass", "normal"},
				Required:    true,
			},
		},
	})

	Register(Command{
		Name:        "save-metadata",
		Description: "Save project metadata (title, description, instructions) for the current cw project.",
		CLICommand:  "save-metadata",
		Internal:    true,
		Params: map[string]ParamDef{
			"json": {
				Type:        "string",
				Description: `JSON string with fields: title, description, instructions. Example: {"title":"My Project","description":"...","instructions":"..."}`,
				Required:    true,
			},
		},
	})

	// Prompt-only commands — plugin body lives in embedded .md files

	Register(Command{
		Name:        "mode",
		Description: "Switch or show the current behavioral mode. Usage: /mode [code|research|review|plan|tdd|debug|free]",
	})

	Register(Command{
		Name:        "permissions",
		Description: "Switch permissions mode. Usage: /permissions [bypass|normal]",
	})

	Register(Command{
		Name:        "help",
		Description: "Show all cw commands and modes available inside Claude.",
	})

	Register(Command{
		Name:        "yolo",
		Description: "Plan and execute work autonomously. Usage: /yolo [objective]",
	})

	Register(Command{
		Name:        "yolo-start",
		Description: "Start yolo autonomous execution. Writes sideband file and triggers reload.",
		CLICommand:  "yolo-start",
		Internal:    true,
	})

	Register(Command{
		Name:        "yolo-stop",
		Description: "Stop yolo autonomous execution. Deletes plan file and triggers reload.",
		CLICommand:  "yolo-stop",
		Internal:    true,
	})

	Register(Command{
		Name:        "new-task",
		Description: "Create a new task: map codebase, understand intent, create worktree branches, and set up task context.",
	})

	Register(Command{
		Name:        "finish-task",
		Description: "Complete the current task: verify clean state, remove worktrees, and return to task selection.",
	})

	Register(Command{
		Name:        "setup-project",
		Description: "Map the codebase and save project metadata for a new cw project.",
	})

	Register(Command{
		Name:        "save-task",
		Description: "Save task metadata and create worktrees for a new task.",
		CLICommand:  "save-task",
		Internal:    true,
		Params: map[string]ParamDef{
			"json": {
				Type:        "string",
				Description: `JSON string with fields: name, description, branch. Example: {"name":"add-auth","description":"...","branch":"feat/add-auth"}`,
				Required:    true,
			},
		},
	})

	Register(Command{
		Name:        "complete-task",
		Description: "Mark task as completed and remove its worktrees.",
		CLICommand:  "finish-task",
		Internal:    true,
	})

	Register(Command{
		Name:        "dev",
		Description: "Run development commands (dev servers, watchers, type generators) in the background. Auto-discovers commands on first run. Usage: /dev [stop|restart|status|update|logs]",
	})
}
