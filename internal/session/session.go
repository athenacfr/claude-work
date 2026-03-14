package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ahtwr/cw/internal/paths"
)

// Session represents a cw-managed session stored in <project>/.cw/sessions/.
// The ID is a UUID that is shared with Claude via --session-id on first launch,
// so both cw and Claude use the same identifier for the session.
type Session struct {
	ID              string `json:"id"`
	Mode            string `json:"mode"`
	SkipPermissions bool   `json:"skip_permissions"`
	Summary         string `json:"summary"`
	StartedAt       string `json:"started_at"`
	LastActive      string `json:"last_active"`
	Status          string `json:"status"` // "active" or "completed"
}

// sessionsDir returns the .cw/sessions/ directory for a project.
func sessionsDir(projectDir string) string {
	return filepath.Join(projectDir, ".cw", "sessions")
}

// sessionPath returns the file path for a session.
func sessionPath(projectDir, id string) string {
	return filepath.Join(sessionsDir(projectDir), id+".json")
}

// New creates a new session with the given parameters.
func New(id, mode string, skipPerms bool) *Session {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Session{
		ID:              id,
		Mode:            mode,
		SkipPermissions: skipPerms,
		StartedAt:       now,
		LastActive:      now,
		Status:          "active",
	}
}

// Save writes the session to <projectDir>/.cw/sessions/<id>.json.
func (s *Session) Save(projectDir string) error {
	dir := sessionsDir(projectDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	return os.WriteFile(sessionPath(projectDir, s.ID), data, 0644)
}

// Load reads a session from disk.
func Load(projectDir, id string) (*Session, error) {
	data, err := os.ReadFile(sessionPath(projectDir, id))
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &s, nil
}

// List returns all sessions for a project, sorted by last_active descending.
func List(projectDir string) ([]Session, error) {
	dir := sessionsDir(projectDir)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []Session
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(f.Name(), ".json")
		s, err := Load(projectDir, id)
		if err != nil {
			continue
		}
		sessions = append(sessions, *s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActive > sessions[j].LastActive
	})

	return sessions, nil
}

// Touch updates the session's LastActive timestamp and saves.
func (s *Session) Touch(projectDir string) error {
	s.LastActive = time.Now().UTC().Format(time.RFC3339)
	return s.Save(projectDir)
}

// Delete removes a session file.
func Delete(projectDir, id string) error {
	return os.Remove(sessionPath(projectDir, id))
}

// ProjectSessionsDir returns the .cw/sessions/ path for a project name.
func ProjectSessionsDir(name string) string {
	return sessionsDir(filepath.Join(paths.ProjectsDir(), name))
}

// RelativeTime returns a human-readable relative time string.
func RelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return itoa(m) + "m ago"
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return itoa(h) + "h ago"
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return itoa(days) + "d ago"
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// ParseTime parses an RFC3339 timestamp string.
func ParseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ExtractSummary reads the Claude JSONL file for a session and extracts
// the first user message as a summary. workDir is the directory Claude
// was launched in (used to locate the JSONL file).
func ExtractSummary(sessionID, workDir string) string {
	jsonlPath := claudeJSONLPath(sessionID, workDir)
	if jsonlPath == "" {
		return ""
	}

	f, err := os.Open(jsonlPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		var rec struct {
			Type    string `json:"type"`
			Message struct {
				Role    string `json:"role"`
				Content any    `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &rec) != nil {
			continue
		}

		if rec.Type != "user" || rec.Message.Role != "user" {
			continue
		}

		text := extractMessageText(rec.Message.Content)
		text = strings.TrimSpace(text)

		// Skip command/system messages
		if text == "" || strings.HasPrefix(text, "<") {
			continue
		}

		// Truncate to a reasonable length
		if utf8.RuneCountInString(text) > 120 {
			runes := []rune(text)
			text = string(runes[:120])
		}

		return text
	}

	return ""
}

// claudeJSONLPath returns the path to Claude's JSONL file for a session.
// Claude stores sessions at ~/.claude/projects/<encoded-workdir>/<session-id>.jsonl
// where encoded-workdir has "/" replaced by "-".
func claudeJSONLPath(sessionID, workDir string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	encoded := strings.ReplaceAll(workDir, string(filepath.Separator), "-")
	path := filepath.Join(home, ".claude", "projects", encoded, sessionID+".jsonl")

	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

// extractMessageText gets the text content from a user message.
// Content can be a string or an array of content blocks.
func extractMessageText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		for _, block := range v {
			if m, ok := block.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					return t
				}
			}
		}
	}
	return ""
}
