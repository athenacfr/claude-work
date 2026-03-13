package tui

import (
	"github.com/ahtwr/cw/internal/claude"
	"github.com/ahtwr/cw/internal/git"
	"github.com/ahtwr/cw/internal/project"
	"github.com/ahtwr/cw/internal/tui/shared"
	"github.com/ahtwr/cw/internal/tui/screen"
	"github.com/ahtwr/cw/internal/tui/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type activeScreen int

const (
	screenProjectList activeScreen = iota
	screenCreateProject
	screenEditProject
	screenModeSelect
)

type Model struct {
	screen activeScreen
	width  int
	height int

	selectedProject *project.Project
	launchConfig    *claude.LaunchConfig

	projectList   screen.ProjectListModel
	createProject screen.CreateProjectModel
	editProject   screen.EditProjectModel
	modeSelect    screen.ModeSelectModel
}

func NewModel(pluginDir string) Model {
	return Model{
		screen:      screenProjectList,
		projectList: screen.NewProjectListModel(),
		launchConfig: &claude.LaunchConfig{
			PluginDir: pluginDir,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return m.projectList.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.projectList.SetSize(msg.Width, msg.Height)
		m.createProject.SetSize(msg.Width, msg.Height)
		m.editProject.SetSize(msg.Width, msg.Height)
		m.modeSelect.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, style.GlobalKeys.Quit) {
			if m.screen == screenCreateProject {
				m.createProject.CleanupIfNeeded()
			}
			m.launchConfig.WorkDir = ""
			return m, tea.Quit
		}

	case shared.RepoSelectedMsg:
		full, err := project.Get(msg.ProjectName)
		if err != nil {
			return m, nil
		}
		for _, r := range full.Repos {
			if r.Name == msg.RepoName {
				m.launchConfig.WorkDir = r.Path
				m.launchConfig.ProjectName = msg.RepoName
				m.launchConfig.EditorMode = true

				go func() {
					ch := git.PullAll(full.Path, []string{msg.RepoName})
					for range ch {
					}
				}()

				return m, tea.Quit
			}
		}
		return m, nil

	case shared.ProjectSelectedMsg:
		m.selectedProject = msg.Project
		m.launchConfig.WorkDir = m.selectedProject.Path
		m.launchConfig.ProjectName = m.selectedProject.Name

		var repoNames []string
		for _, r := range m.selectedProject.Repos {
			repoNames = append(repoNames, r.Name)
		}
		go func() {
			ch := git.PullAll(m.selectedProject.Path, repoNames)
			for range ch {
			}
		}()

		if !project.HasMetadata(msg.Project.Name) {
			m.launchConfig.Prompt = "/cw:new-intention"
			m.launchConfig.SkipPermissions = true
			m.launchConfig.AutoSetup = true
			m.launchConfig.AutoCompactLimit = m.projectList.AutoCompactLimit
			return m, tea.Quit
		}

		m.launchConfig.AutoCompactLimit = m.projectList.AutoCompactLimit

		bypass := m.projectList.BypassPerms
		m.screen = screenModeSelect
		m.modeSelect = screen.NewModeSelectModelWithBypass(bypass, m.selectedProject.Path)
		m.modeSelect.SetSize(m.width, m.height)
		return m, m.modeSelect.LoadSessions(m.selectedProject.Path)

	case shared.ModeSelectedMsg:
		m.launchConfig.Mode = msg.Mode
		m.launchConfig.SkipPermissions = msg.SkipPermissions
		switch msg.SessionKind {
		case 1:
			m.launchConfig.Resume = true
		case 2:
			m.launchConfig.SessionID = msg.SessionID
		}
		return m, tea.Quit

	case shared.NavigateMsg:
		switch msg.Screen {
		case shared.ScreenProjectList:
			m.screen = screenProjectList
			m.projectList = screen.NewProjectListModel()
			m.projectList.SetSize(m.width, m.height)
			return m, m.projectList.Init()
		case shared.ScreenCreateProject:
			m.screen = screenCreateProject
			m.createProject = screen.NewCreateProjectModel()
			m.createProject.SetSize(m.width, m.height)
			return m, m.createProject.Init()
		case shared.ScreenEditProject:
			m.screen = screenEditProject
			m.editProject = screen.NewEditProjectModel(msg.ProjectName)
			m.editProject.SetSize(m.width, m.height)
			if msg.AddRepo {
				m.editProject.Step = screen.EditStepMethod
				m.editProject.MethodList = m.editProject.BuildMethodList()
			}
			return m, m.editProject.Init()
		}
		return m, nil

	case shared.LaunchMsg:
		return m, tea.Quit
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenProjectList:
		m.projectList, cmd = m.projectList.Update(msg)
	case screenCreateProject:
		m.createProject, cmd = m.createProject.Update(msg)
	case screenEditProject:
		m.editProject, cmd = m.editProject.Update(msg)
	case screenModeSelect:
		m.modeSelect, cmd = m.modeSelect.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.screen {
	case screenProjectList:
		return m.projectList.View()
	case screenCreateProject:
		return m.createProject.View()
	case screenEditProject:
		return m.editProject.View()
	case screenModeSelect:
		return m.modeSelect.View()
	}
	return ""
}

func (m Model) ShouldLaunch() bool {
	return m.launchConfig != nil && m.launchConfig.WorkDir != ""
}

func (m Model) LaunchConfig() claude.LaunchConfig {
	if m.launchConfig != nil {
		return *m.launchConfig
	}
	return claude.LaunchConfig{}
}
