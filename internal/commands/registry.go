package commands

// Command defines a cw command with its metadata for plugin generation.
type Command struct {
	Name        string
	Description string
	Params      map[string]ParamDef // nil means no params
	CLICommand  string              // internal CLI subcommand (e.g. "reload", "mode-switch"); empty for prompt-only commands
	PluginBody  string              // custom .md body for plugin generation; if empty, auto-generated from CLICommand
}

// ParamDef describes a single command parameter.
type ParamDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"-"`
}

var commands []Command

// Register adds a command to the global registry.
func Register(c Command) {
	commands = append(commands, c)
}

// All returns all registered commands.
func All() []Command {
	return commands
}
