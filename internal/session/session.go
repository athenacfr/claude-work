package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Session represents a Claude Code conversation session.
type Session struct {
	ID        string
	Timestamp time.Time
	Summary   string // first real user message, truncated
	SizeKB    int64  // file size in KB
}

// List discovers sessions for a given project directory by reading
// Claude Code's session storage (~/.claude/projects/<encoded-path>/).
// Returns sessions sorted by timestamp descending (most recent first).
// Returns nil, nil if no sessions found or directory doesn't exist.
func List(projectDir string) ([]Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	encoded := strings.ReplaceAll(projectDir, "/", "-")
	sessionDir := filepath.Join(home, ".claude", "projects", encoded)

	files, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []Session
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
			continue
		}

		id := strings.TrimSuffix(f.Name(), ".jsonl")
		fullPath := filepath.Join(sessionDir, f.Name())

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		sizeKB := info.Size() / 1024

		s, err := parseSession(fullPath, id)
		if err != nil || s == nil {
			continue
		}
		s.SizeKB = sizeKB
		sessions = append(sessions, *s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp.After(sessions[j].Timestamp)
	})

	return sessions, nil
}

// parseSession reads a .jsonl file and extracts the first real user message.
func parseSession(path, id string) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)

	var firstTS time.Time

	for scanner.Scan() {
		var entry struct {
			Type    string `json:"type"`
			Message struct {
				Role    string `json:"role"`
				Content any    `json:"content"`
			} `json:"message"`
			Timestamp string `json:"timestamp"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if entry.Type != "user" || entry.Message.Role != "user" {
			continue
		}

		text := extractText(entry.Message.Content)
		if text == "" {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if firstTS.IsZero() {
			firstTS = ts
		}

		summary := cleanSummary(text, 200)
		if summary == "" {
			continue
		}

		return &Session{
			ID:        id,
			Timestamp: firstTS,
			Summary:   summary,
		}, nil
	}

	// Session exists but no readable user message
	if !firstTS.IsZero() {
		return &Session{
			ID:        id,
			Timestamp: firstTS,
			Summary:   "(no summary)",
		}, nil
	}

	return nil, nil
}

// extractText gets a string from the message content (can be string or array).
func extractText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					return t
				}
			}
		}
	}
	return ""
}

// systemPrefixes are content prefixes that indicate system-injected messages, not real user input.
var systemPrefixes = []string{
	"<local-command-caveat>",
	"<system-reminder>",
	"<command-message>",
	"<command-name>",
	"<command-args>",
	"<local-command-stdout>",
}

// cleanSummary strips XML-like tags and truncates.
// Returns empty string if the content is purely system-injected.
func cleanSummary(s string, maxLen int) string {
	trimmed := strings.TrimSpace(s)

	// Skip messages that start with system tags
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return ""
		}
	}

	// Strip all <tag>...</tag> and standalone <tag> patterns
	text := stripTags(trimmed)
	text = strings.TrimSpace(text)

	if text == "" {
		return ""
	}

	// Collapse whitespace
	parts := strings.Fields(text)
	text = strings.Join(parts, " ")

	if len(text) > maxLen {
		text = text[:maxLen] + "..."
	}
	return text
}

// stripTags removes all XML-like tags from a string.
func stripTags(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '<' {
			end := strings.IndexByte(s[i:], '>')
			if end != -1 {
				i += end + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// message holds a role + text pair extracted from a session JSONL.
type message struct {
	Role string
	Text string
}

// ExtractRecentContext reads the most recent session JSONL for projectDir
// and returns the last maxMessages user/assistant exchange pairs formatted
// as a continuation prompt. Returns "" if no context is available.
func ExtractRecentContext(projectDir string, maxMessages int) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	encoded := strings.ReplaceAll(projectDir, "/", "-")
	sessionDir := filepath.Join(home, ".claude", "projects", encoded)

	// Find the most recently modified .jsonl file
	files, err := os.ReadDir(sessionDir)
	if err != nil {
		return ""
	}

	var newest string
	var newestTime time.Time
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newestTime) {
			newestTime = info.ModTime()
			newest = filepath.Join(sessionDir, f.Name())
		}
	}
	if newest == "" {
		return ""
	}

	messages := extractMessages(newest, maxMessages)
	if len(messages) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Auto-compact was triggered. Continue working on what you were doing. Here's the recent context:\n\n")
	for _, m := range messages {
		if m.Role == "user" {
			b.WriteString("User: ")
		} else {
			b.WriteString("Assistant: ")
		}
		b.WriteString(m.Text)
		b.WriteString("\n\n")
	}
	return b.String()
}

// extractMessages reads a session JSONL and returns the last N user+assistant
// text messages, skipping system-injected and tool_result/tool_use entries.
func extractMessages(path string, maxPairs int) []message {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var msgs []message

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var entry struct {
			Type    string `json:"type"`
			Message struct {
				Role    string `json:"role"`
				Content any    `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		switch entry.Type {
		case "user":
			if text := extractUserText(entry.Message.Content); text != "" {
				cleaned := cleanSummary(text, 500)
				if cleaned != "" {
					msgs = append(msgs, message{Role: "user", Text: cleaned})
				}
			}
		case "assistant":
			if text := extractAssistantText(entry.Message.Content); text != "" {
				truncated := text
				if len(truncated) > 500 {
					truncated = truncated[:500] + "..."
				}
				msgs = append(msgs, message{Role: "assistant", Text: truncated})
			}
		}
	}

	// Keep only the last N pairs (up to maxPairs*2 messages)
	limit := maxPairs * 2
	if len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}
	return msgs
}

// extractUserText gets text from a user message, skipping tool_result entries.
func extractUserText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if m["type"] == "text" {
				if t, ok := m["text"].(string); ok {
					return t
				}
			}
		}
	}
	return ""
}

// extractAssistantText collects text blocks from an assistant message,
// skipping tool_use blocks.
func extractAssistantText(content any) string {
	items, ok := content.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if m["type"] == "text" {
			if t, ok := m["text"].(string); ok && strings.TrimSpace(t) != "" {
				parts = append(parts, strings.TrimSpace(t))
			}
		}
	}
	return strings.Join(parts, " ")
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
