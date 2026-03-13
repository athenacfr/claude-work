package screen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahtwr/cw/internal/gh"
	"github.com/ahtwr/cw/internal/git"
	"github.com/ahtwr/cw/internal/paths"
	"github.com/ahtwr/cw/internal/project"
	"github.com/ahtwr/cw/internal/tui/shared"
	"github.com/ahtwr/cw/internal/tui/style"
	"github.com/ahtwr/cw/internal/tui/widget"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type editStep int

const (
	editStepMain editStep = iota
	editStepRename
	EditStepMethod // exported — app.go sets this directly
	editStepRepos
	editStepGitURL
	editStepLocalPath
	editStepCopyMove
	editStepCloning
	editStepConfirmRemove
)

type editRepoItem struct {
	name       string
	branch     string
	dirtyCount int
}

func (e editRepoItem) FilterValue() string { return e.name }

type repoRemovedMsg struct{}
type projectRenamedMsg struct{ newName string }

type EditProjectModel struct {
	Step        editStep
	width       int
	height      int
	projectName string
	projectPath string
	repoList    widget.FzfListModel

	renameInput textinput.Model
	renameErr   string

	removeTarget string

	MethodList   widget.FzfListModel
	ghRepoList   widget.FzfListModel
	urlInput     textinput.Model
	pathInput    textinput.Model
	copyMoveList widget.FzfListModel
	localPath    string
	spinner      spinner.Model
	statusText   string
	progress     widget.ProgressModel
	cloneChan    <-chan git.CloneProgress
	loading      bool
	ghAvail      bool
	reposErr     error

	errMsg string
}

func NewEditProjectModel(name string) EditProjectModel {
	urlTi := textinput.New()
	urlTi.Placeholder = "https://github.com/user/repo.git"
	urlTi.CharLimit = 200

	pathTi := textinput.New()
	pathTi.Placeholder = "/path/to/directory"
	pathTi.CharLimit = 200

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	projPath := filepath.Join(paths.ProjectsDir(), name)

	m := EditProjectModel{
		Step:        editStepMain,
		projectName: name,
		projectPath: projPath,
		urlInput:    urlTi,
		pathInput:   pathTi,
		spinner:     sp,
		ghAvail:     gh.IsAvailable() && gh.IsAuthenticated(),
	}
	m.refreshRepoList()
	return m
}

func (m *EditProjectModel) refreshRepoList() {
	p, err := project.Get(m.projectName)
	if err != nil {
		return
	}
	m.projectPath = p.Path
	items := make([]widget.FzfItem, len(p.Repos))
	for i, r := range p.Repos {
		items[i] = editRepoItem{
			name:       r.Name,
			branch:     r.Branch,
			dirtyCount: len(r.DirtyFiles),
		}
	}
	m.repoList = widget.NewFzfList(items, widget.FzfListConfig{
		RenderItem:   renderEditRepoItem,
		PreviewFunc:  editRepoPreview(m.projectPath),
		Placeholder:  "No repos. Press 'a' to add one.",
		ListWidthPct: 0.4,
	})
	m.repoList.SetSize(m.width, m.height-5)
}

func (m EditProjectModel) Init() tea.Cmd {
	return nil
}

func (m *EditProjectModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.repoList.SetSize(w, h-5)
	m.MethodList.SetSize(w, h-5)
	m.ghRepoList.SetSize(w, h-5)
	m.copyMoveList.SetSize(w, h-5)
}

func (m EditProjectModel) BuildMethodList() widget.FzfListModel {
	var items []widget.FzfItem
	if m.ghAvail {
		items = append(items, shared.MethodItem{Name: "GitHub"})
	}
	items = append(items, shared.MethodItem{Name: "Git URL"})
	items = append(items, shared.MethodItem{Name: "Local directory"})
	items = append(items, shared.MethodItem{Name: "Empty (git init)"})
	list := widget.NewFzfList(items, widget.FzfListConfig{
		Placeholder: "No methods",
	})
	list.SetSize(m.width, m.height-5)
	return list
}

func (m EditProjectModel) Update(msg tea.Msg) (EditProjectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case shared.AllReposLoadedMsg:
		m.loading = false
		m.reposErr = msg.Err
		if msg.Err == nil {
			m.ghRepoList = widget.NewFzfList(msg.Repos, widget.FzfListConfig{
				MultiSelect:  true,
				PreviewFunc:  shared.RepoPreview,
				RenderItem:   shared.RenderRepoItem,
				Placeholder:  "No repos found",
				ListWidthPct: 0.45,
			})
			m.ghRepoList.SetSize(m.width, m.height-5)
		}
		return m, nil

	case widget.CloneTickMsg:
		m.progress, _ = m.progress.Update(msg)
		return m, shared.ListenForCloneProgress(m.cloneChan)

	case widget.CloneAllDoneMsg:
		m.refreshRepoList()
		m.Step = editStepMain
		return m, nil

	case shared.AddCompleteMsg:
		m.loading = false
		m.refreshRepoList()
		m.Step = editStepMain
		return m, nil

	case repoRemovedMsg:
		m.refreshRepoList()
		m.Step = editStepMain
		return m, nil

	case projectRenamedMsg:
		m.projectName = msg.newName
		m.projectPath = filepath.Join(paths.ProjectsDir(), msg.newName)
		m.Step = editStepMain
		return m, nil

	case spinner.TickMsg:
		if m.Step == editStepCloning {
			if m.cloneChan != nil {
				var cmd tea.Cmd
				m.progress, cmd = m.progress.Update(msg)
				return m, cmd
			}
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		m.errMsg = ""

		switch m.Step {
		case editStepMain:
			return m.updateMain(msg)
		case editStepRename:
			return m.updateRename(msg)
		case editStepConfirmRemove:
			return m.updateConfirmRemove(msg)
		case EditStepMethod:
			return m.updateMethod(msg)
		case editStepRepos:
			return m.updateGHRepos(msg)
		case editStepGitURL:
			return m.updateGitURL(msg)
		case editStepLocalPath:
			return m.updateLocalPath(msg)
		case editStepCopyMove:
			return m.updateCopyMove(msg)
		}
	}
	return m, nil
}

func (m EditProjectModel) updateMain(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	newList, consumed, result := m.repoList.HandleKey(msg.String())
	m.repoList = newList

	if result != nil {
		switch result.(type) {
		case widget.FzfCancelMsg:
			return m, func() tea.Msg {
				return shared.NavigateMsg{Screen: shared.ScreenProjectList}
			}
		}
	}

	if consumed {
		return m, nil
	}

	if !m.repoList.IsSearching() {
		switch msg.String() {
		case "a":
			m.Step = EditStepMethod
			m.MethodList = m.BuildMethodList()
			return m, nil
		case "x", "d":
			item := m.repoList.SelectedItem()
			if item != nil {
				ri := item.(editRepoItem)
				m.removeTarget = ri.name
				m.Step = editStepConfirmRemove
			}
			return m, nil
		case "r":
			m.Step = editStepRename
			m.renameErr = ""
			ti := textinput.New()
			ti.SetValue(m.projectName)
			ti.Focus()
			ti.CharLimit = 50
			m.renameInput = ti
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m EditProjectModel) updateRename(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		newName := strings.TrimSpace(m.renameInput.Value())
		if newName == "" {
			m.renameErr = "Name cannot be empty"
			return m, nil
		}
		if newName == m.projectName {
			m.Step = editStepMain
			return m, nil
		}
		newPath := filepath.Join(paths.ProjectsDir(), newName)
		if _, err := os.Stat(newPath); err == nil {
			m.renameErr = "A project with that name already exists"
			return m, nil
		}
		if err := project.Rename(m.projectName, newName); err != nil {
			m.renameErr = "Rename failed: " + err.Error()
			return m, nil
		}
		m.projectName = newName
		m.projectPath = newPath
		m.Step = editStepMain
		return m, nil
	case "esc":
		m.Step = editStepMain
		return m, nil
	default:
		m.renameErr = ""
		var cmd tea.Cmd
		m.renameInput, cmd = m.renameInput.Update(msg)
		return m, cmd
	}
}

func (m EditProjectModel) updateConfirmRemove(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		name := m.removeTarget
		m.removeTarget = ""
		return m, func() tea.Msg {
			project.RemoveRepo(m.projectName, name)
			return repoRemovedMsg{}
		}
	default:
		m.removeTarget = ""
		m.Step = editStepMain
		return m, nil
	}
}

func (m EditProjectModel) updateMethod(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	newList, consumed, result := m.MethodList.HandleKey(msg.String())
	m.MethodList = newList

	if result != nil {
		switch r := result.(type) {
		case widget.FzfConfirmMsg:
			mi := r.Item.(shared.MethodItem)
			switch mi.Name {
			case "GitHub":
				m.Step = editStepRepos
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, shared.LoadAllRepos)
			case "Git URL":
				m.Step = editStepGitURL
				m.urlInput.SetValue("")
				m.urlInput.Focus()
				return m, textinput.Blink
			case "Local directory":
				m.Step = editStepLocalPath
				m.pathInput.SetValue("")
				m.pathInput.Focus()
				return m, textinput.Blink
			case "Empty (git init)":
				m.Step = editStepCloning
				m.statusText = "Initializing repo..."
				return m, tea.Batch(
					m.spinner.Tick,
					initEmptyRepoInProject(m.projectName),
				)
			}
		case widget.FzfCancelMsg:
			m.Step = editStepMain
			return m, nil
		}
	}

	if consumed {
		return m, nil
	}
	return m, nil
}

func (m EditProjectModel) updateGHRepos(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	newList, consumed, result := m.ghRepoList.HandleKey(msg.String())
	m.ghRepoList = newList

	if result != nil {
		switch r := result.(type) {
		case widget.FzfConfirmMsg:
			var selected []shared.RepoItem
			for _, item := range r.Items {
				ri := item.(shared.RepoItem)
				selected = append(selected, ri)
			}
			if len(selected) == 0 {
				return m, nil
			}

			projDir := filepath.Join(paths.ProjectsDir(), m.projectName)
			var repoNames []string
			for _, r := range selected {
				name := r.NameWithOwner
				if name == "" {
					name = r.Org + "/" + r.Name
				}
				repoNames = append(repoNames, name)
			}

			ch := git.ParallelClone(projDir, repoNames)
			m.cloneChan = ch
			m.Step = editStepCloning
			m.progress = widget.NewProgressModel("Cloning repos", repoNames)
			m.progress.SetSize(m.width, m.height)

			return m, tea.Batch(
				m.progress.Init(),
				shared.ListenForCloneProgress(ch),
			)
		case widget.FzfCancelMsg:
			m.Step = EditStepMethod
			m.MethodList = m.BuildMethodList()
			return m, nil
		}
	}

	if consumed {
		return m, nil
	}
	return m, nil
}

func (m EditProjectModel) updateGitURL(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		url := strings.TrimSpace(m.urlInput.Value())
		if url == "" {
			return m, nil
		}
		m.Step = editStepCloning
		m.statusText = "Cloning..."
		return m, tea.Batch(
			m.spinner.Tick,
			cloneURLToProject(m.projectName, url),
		)
	case "esc":
		m.Step = EditStepMethod
		m.MethodList = m.BuildMethodList()
		return m, nil
	default:
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}
}

func (m EditProjectModel) updateLocalPath(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		p := strings.TrimSpace(m.pathInput.Value())
		if p == "" {
			return m, nil
		}
		if strings.HasPrefix(p, "~/") {
			home, _ := os.UserHomeDir()
			p = filepath.Join(home, p[2:])
		}
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			m.errMsg = "Directory not found"
			return m, nil
		}
		m.localPath = p
		m.Step = editStepCopyMove
		items := []widget.FzfItem{
			shared.CopyMoveItem{Name: "Copy"},
			shared.CopyMoveItem{Name: "Move"},
		}
		m.copyMoveList = widget.NewFzfList(items, widget.FzfListConfig{
			Placeholder: "Choose action",
		})
		m.copyMoveList.SetSize(m.width, m.height-5)
		return m, nil
	case "esc":
		m.Step = EditStepMethod
		m.MethodList = m.BuildMethodList()
		return m, nil
	default:
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}
}

func (m EditProjectModel) updateCopyMove(msg tea.KeyMsg) (EditProjectModel, tea.Cmd) {
	newList, consumed, result := m.copyMoveList.HandleKey(msg.String())
	m.copyMoveList = newList

	if result != nil {
		switch r := result.(type) {
		case widget.FzfConfirmMsg:
			ci := r.Item.(shared.CopyMoveItem)
			move := ci.Name == "Move"
			action := "Copying..."
			if move {
				action = "Moving..."
			}
			m.Step = editStepCloning
			m.statusText = action
			return m, tea.Batch(
				m.spinner.Tick,
				addLocalDirToProject(m.projectName, m.localPath, move),
			)
		case widget.FzfCancelMsg:
			m.Step = editStepLocalPath
			m.pathInput.Focus()
			return m, textinput.Blink
		}
	}

	if consumed {
		return m, nil
	}
	return m, nil
}

func cloneURLToProject(projectName, url string) tea.Cmd {
	return func() tea.Msg {
		projDir := filepath.Join(paths.ProjectsDir(), projectName)
		name := filepath.Base(strings.TrimSuffix(url, ".git"))
		if name == "" || name == "." {
			name = "repo"
		}
		git.Clone(url, filepath.Join(projDir, name))
		return shared.AddCompleteMsg{}
	}
}

func addLocalDirToProject(projectName, srcPath string, move bool) tea.Cmd {
	return func() tea.Msg {
		projDir := filepath.Join(paths.ProjectsDir(), projectName)
		name := filepath.Base(srcPath)
		dest := filepath.Join(projDir, name)
		if move {
			project.MoveDir(srcPath, dest)
		} else {
			project.CopyDir(srcPath, dest)
		}
		return shared.AddCompleteMsg{}
	}
}

func initEmptyRepoInProject(projectName string) tea.Cmd {
	return func() tea.Msg {
		projDir := filepath.Join(paths.ProjectsDir(), projectName)
		git.Init(projDir)
		return shared.AddCompleteMsg{}
	}
}

// Views

func (m EditProjectModel) View() string {
	switch m.Step {
	case editStepMain:
		return m.viewMain()
	case editStepRename:
		return m.viewRename()
	case editStepConfirmRemove:
		return m.viewConfirmRemove()
	case EditStepMethod:
		return m.viewMethod()
	case editStepRepos:
		return m.viewGHRepos()
	case editStepGitURL:
		return m.viewGitURL()
	case editStepLocalPath:
		return m.viewLocalPath()
	case editStepCopyMove:
		return m.viewCopyMove()
	case editStepCloning:
		return m.viewCloning()
	}
	return ""
}

func (m EditProjectModel) viewMain() string {
	repoCount := m.repoList.TotalCount()
	countStr := fmt.Sprintf("%d repo", repoCount)
	if repoCount != 1 {
		countStr += "s"
	}

	header := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName) +
		"  " + style.DimStyle.Render("("+countStr+")") + "\n"
	kb := style.RenderKeybar(
		style.KeyBind{Key: "a", Desc: "add repo"},
		style.KeyBind{Key: "x", Desc: "remove"},
		style.KeyBind{Key: "r", Desc: "rename"},
		style.KeyBind{Key: "s", Desc: "search"},
		style.KeyBind{Key: "esc", Desc: "back"},
	)
	return header + kb + "\n\n" + m.repoList.View()
}

func (m EditProjectModel) viewRename() string {
	header := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName+" › rename") + "\n"
	kb := style.RenderKeybar(style.KeyBind{Key: "enter", Desc: "confirm"}, style.KeyBind{Key: "esc", Desc: "cancel"})
	body := "\n  New name:\n\n  " + m.renameInput.View()
	if m.renameErr != "" {
		body += "\n\n  " + style.ErrorStyle.Render(m.renameErr)
	}
	return header + kb + "\n" + body
}

func (m EditProjectModel) viewConfirmRemove() string {
	header := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName) + "\n"
	warning := "\n" + style.ErrorStyle.Render("  Remove '"+m.removeTarget+"'?") + "\n" +
		style.DimStyle.Render("  This will delete the repo directory from this project.") + "\n\n" +
		"  " + style.KeyStyle.Render("y") + style.DimStyle.Render(" confirm  ") +
		style.KeyStyle.Render("n") + style.DimStyle.Render(" cancel")
	return header + warning
}

func (m EditProjectModel) viewMethod() string {
	breadcrumb := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName+" › add repo") + "\n"
	kb := style.RenderKeybar(style.KeyBind{Key: "enter", Desc: "select"}, style.KeyBind{Key: "esc", Desc: "back"})
	return breadcrumb + kb + "\n\n" + m.MethodList.View()
}

func (m EditProjectModel) viewGHRepos() string {
	breadcrumb := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName+" › GitHub") + "\n"

	if m.loading {
		return breadcrumb + fmt.Sprintf("\n  %s Loading repos...", m.spinner.View())
	}
	if m.reposErr != nil {
		kb := style.RenderKeybar(style.KeyBind{Key: "esc", Desc: "back"})
		return breadcrumb + "\n" + style.ErrorStyle.Render("  "+m.reposErr.Error()) + "\n\n" + kb
	}

	kb := style.RenderKeybar(
		style.KeyBind{Key: "space", Desc: "toggle"},
		style.KeyBind{Key: "enter", Desc: "clone"},
		style.KeyBind{Key: "s", Desc: "search"},
		style.KeyBind{Key: "esc", Desc: "back"},
	)
	return breadcrumb + kb + "\n\n" + m.ghRepoList.View()
}

func (m EditProjectModel) viewGitURL() string {
	breadcrumb := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName+" › Git URL") + "\n"
	kb := style.RenderKeybar(style.KeyBind{Key: "enter", Desc: "clone"}, style.KeyBind{Key: "esc", Desc: "back"})
	body := "\n  Repository URL:\n\n  " + m.urlInput.View()
	return breadcrumb + kb + "\n" + body
}

func (m EditProjectModel) viewLocalPath() string {
	breadcrumb := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName+" › local") + "\n"
	kb := style.RenderKeybar(style.KeyBind{Key: "enter", Desc: "next"}, style.KeyBind{Key: "esc", Desc: "back"})
	body := "\n  Directory path:\n\n  " + m.pathInput.View()
	if m.errMsg != "" {
		body += "\n\n  " + style.ErrorStyle.Render(m.errMsg)
	}
	return breadcrumb + kb + "\n" + body
}

func (m EditProjectModel) viewCopyMove() string {
	breadcrumb := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName+" › "+filepath.Base(m.localPath)) + "\n"
	kb := style.RenderKeybar(style.KeyBind{Key: "enter", Desc: "select"}, style.KeyBind{Key: "esc", Desc: "back"})
	return breadcrumb + kb + "\n\n" + m.copyMoveList.View()
}

func (m EditProjectModel) viewCloning() string {
	if m.cloneChan != nil {
		return m.progress.View()
	}
	header := style.TitleStyle.Render("EDIT PROJECT") +
		style.DimStyle.Render(" › "+m.projectName) + "\n"
	body := fmt.Sprintf("\n  %s %s", m.spinner.View(), m.statusText)
	return header + body
}

func renderEditRepoItem(item widget.FzfItem, index int, cursor, selected bool, matched []int, width int) string {
	ri := item.(editRepoItem)

	prefix := "  "
	if cursor {
		prefix = style.FzfCursorPrefix.Render("▶ ")
	}

	name := widget.HighlightMatches(ri.name, matched)

	meta := style.DimStyle.Render(" on ") + ri.branch
	if ri.dirtyCount > 0 {
		meta += " " + style.ErrorStyle.Render("●")
	} else {
		meta += " " + style.SuccessStyle.Render("✓")
	}

	return prefix + name + meta
}

func editRepoPreview(projectPath string) func(widget.FzfItem, int, int) string {
	return func(item widget.FzfItem, width, height int) string {
		ri := item.(editRepoItem)
		info := git.GetRepoInfo(projectPath, ri.name)

		var lines []string
		lines = append(lines, style.TitleStyle.Render(ri.name))
		lines = append(lines, style.DimStyle.Render("on ")+info.Branch)
		lines = append(lines, "")

		if info.Ahead > 0 || info.Behind > 0 {
			ab := ""
			if info.Ahead > 0 {
				ab += style.SuccessStyle.Render(fmt.Sprintf("↑%d ahead", info.Ahead))
			}
			if info.Behind > 0 {
				if ab != "" {
					ab += "  "
				}
				ab += style.ErrorStyle.Render(fmt.Sprintf("↓%d behind", info.Behind))
			}
			lines = append(lines, ab)
			lines = append(lines, "")
		}

		if info.Clean {
			lines = append(lines, style.SuccessStyle.Render("✓ clean"))
		} else {
			lines = append(lines, style.ErrorStyle.Render(fmt.Sprintf("● %d modified", ri.dirtyCount)))
		}
		lines = append(lines, "")

		if len(info.RecentCommits) > 0 {
			lines = append(lines, style.DimStyle.Render("Recent commits:"))
			for _, commit := range info.RecentCommits {
				lines = append(lines, "  "+style.DimStyle.Render(commit))
			}
		}

		if len(lines) > height {
			lines = lines[:height]
		}
		return strings.Join(lines, "\n")
	}
}
