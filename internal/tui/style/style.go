package style

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	SubtitleStyle = lipgloss.NewStyle().Faint(true)
	KeyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	DimStyle      = lipgloss.NewStyle().Faint(true)
	AccentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	ErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	SuccessStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	// fzf-specific styles
	FzfPromptStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	FzfSearchInput    = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	FzfSearchActive   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	FzfMatchStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	FzfSelectedLine   = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	FzfCounterStyle   = lipgloss.NewStyle().Faint(true)
	FzfBorderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	FzfMarkerSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	FzfMarkerNormal   = lipgloss.NewStyle().Faint(true)
	FzfCursorPrefix   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	FzfDividerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)

	// Tree view styles
	TreeBranchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	TreeAddStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	PreviewStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// KeyBind represents a key binding for the keybar.
type KeyBind struct {
	Key  string
	Desc string
}

// RenderKeybar renders a horizontal key binding bar.
func RenderKeybar(bindings ...KeyBind) string {
	var parts []string
	for _, b := range bindings {
		parts = append(parts, KeyStyle.Render(b.Key)+" "+DimStyle.Render(b.Desc))
	}
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += "  "
		}
		s += p
	}
	return s
}
