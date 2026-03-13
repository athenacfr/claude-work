package widget

import (
	"fmt"
	"strings"

	"github.com/ahtwr/cw/internal/git"
	"github.com/ahtwr/cw/internal/tui/style"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// CloneTickMsg wraps a single clone progress event.
type CloneTickMsg git.CloneProgress

// CloneAllDoneMsg signals all clones are finished.
type CloneAllDoneMsg struct{}

// RepoStatus tracks the clone status of a single repo.
type RepoStatus struct {
	Repo string
	Done bool
	Err  error
}

// ProgressModel shows clone/operation progress for multiple repos.
type ProgressModel struct {
	title   string
	repos   []RepoStatus
	spinner spinner.Model
	done    bool
	width   int
	height  int
}

// NewProgressModel creates a progress model with the given title and repo names.
func NewProgressModel(title string, repoNames []string) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot

	repos := make([]RepoStatus, len(repoNames))
	for i, r := range repoNames {
		repos[i] = RepoStatus{Repo: r}
	}

	return ProgressModel{
		title:   title,
		repos:   repos,
		spinner: s,
	}
}

func (m *ProgressModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case CloneTickMsg:
		for i := range m.repos {
			if m.repos[i].Repo == msg.Repo && msg.Done {
				m.repos[i].Done = true
				m.repos[i].Err = msg.Err
				break
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ProgressModel) View() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s\n\n", style.TitleStyle.Render(m.title)))

	doneCount := 0
	for _, r := range m.repos {
		prefix := m.spinner.View()
		status := style.DimStyle.Render("cloning...")

		if r.Done {
			doneCount++
			if r.Err != nil {
				prefix = style.ErrorStyle.Render("✗")
				status = style.ErrorStyle.Render(r.Err.Error())
			} else {
				prefix = style.SuccessStyle.Render("✓")
				status = style.SuccessStyle.Render("done")
			}
		}

		sb.WriteString(fmt.Sprintf("  %s %s  %s\n", prefix, r.Repo, status))
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n",
		style.DimStyle.Render(fmt.Sprintf("%d/%d complete", doneCount, len(m.repos))),
	))

	return sb.String()
}
