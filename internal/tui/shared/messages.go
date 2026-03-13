package shared

import (
	"github.com/ahtwr/cw/internal/config"
	"github.com/ahtwr/cw/internal/project"
)

// Screen identifies a TUI screen.
type Screen int

const (
	ScreenProjectList Screen = iota
	ScreenCreateProject
	ScreenEditProject
	ScreenModeSelect
)

// ProjectSelectedMsg is sent when a project is selected from the project list.
type ProjectSelectedMsg struct{ Project *project.Project }

// RepoSelectedMsg is sent when a repo is directly selected.
type RepoSelectedMsg struct {
	ProjectName string
	RepoName    string
}

// ModeSelectedMsg is sent when a mode and session are confirmed.
type ModeSelectedMsg struct {
	Mode            config.Mode
	SkipPermissions bool
	SessionKind     int
	SessionID       string
}

// NavigateMsg is sent to switch between screens.
type NavigateMsg struct {
	Screen      Screen
	ProjectName string
	AddRepo     bool
}

// LaunchMsg signals that Claude should be launched.
type LaunchMsg struct{}
