package project

import (
	"fmt"
	"strings"
)

// systemPromptTemplate is the static system prompt bundled with cw.
// The only dynamic part is the subproject list, which is injected at launch time.
const multiRepoTemplate = `# CW Project Context

You are working in a cw-managed project directory that contains multiple subprojects. Each subproject is an independent git repository in its own subfolder.

## Subprojects

%s

## Rules

- Do NOT create code files, configs, or dependencies in the project root directory.
- Each subproject is an independent git repo with its own git history, dependencies, and configuration.
- When working on code, always operate within the appropriate subproject subfolder.
- Do not mix concerns between subprojects — treat each as isolated unless the user explicitly asks for cross-project work.
- The .claude/CLAUDE.md contains project-wide instructions and conventions.
- Never modify the root ` + "`.claude/`" + ` directory. To add or update rules, commands, or settings, do it inside the appropriate subproject's ` + "`.claude/`" + ` directory.

## Working Across Subprojects

When a task spans multiple subprojects:
- Identify which subprojects are involved before making changes.
- Make changes in each subproject independently — commit separately, test separately.
- If subprojects depend on each other (e.g., shared types, API contracts), change the dependency first.
- Use the Agent tool to work on independent subprojects in parallel when possible.
`

const singleRepoTemplate = `# CW Project Context

You are working in a cw-managed project directory with a single subproject.

## Subproject

%s

## Rules

- Do NOT create code files, configs, or dependencies in the project root directory.
- The subproject subfolder is an independent git repo with its own git history, dependencies, and configuration.
- When working on code, always operate within the subproject subfolder.
- The .claude/CLAUDE.md contains project-wide instructions and conventions.
- Never modify the root ` + "`.claude/`" + ` directory. To add or update rules, commands, or settings, do it inside the subproject's ` + "`.claude/`" + ` directory.
`

const agentEncouragement = `
## Agents

You have access to specialized agents via the Agent tool. Use them when beneficial:
- **researcher** — Read-only code exploration and analysis
- **implementer** — Focused implementation of a specific task
- **tester** — Write and run tests
- **reviewer** — Review code changes for quality and bugs

Prefer agents for:
- Parallel independent tasks
- Focused work that benefits from clean context
- Tasks where the agent's specialization matches the work
`

// BuildSystemPrompt returns the system prompt string for a project.
func BuildSystemPrompt(name string) (string, error) {
	p, err := Get(name)
	if err != nil {
		return "", err
	}

	var repoList []string
	for _, r := range p.Repos {
		repoList = append(repoList, fmt.Sprintf("- `%s/`", r.Name))
	}

	tmpl := multiRepoTemplate
	if len(p.Repos) == 1 {
		tmpl = singleRepoTemplate
	}

	return fmt.Sprintf(tmpl, strings.Join(repoList, "\n")) + agentEncouragement, nil
}


