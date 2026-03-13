package config

import (
	"testing"
)

func TestInitModes(t *testing.T) {
	InitModes("/tmp/modes")

	if len(Modes) == 0 {
		t.Fatal("expected modes to be initialized")
	}

	// Should contain the standard modes
	names := make(map[string]bool)
	for _, m := range Modes {
		names[m.Name] = true
	}

	required := []string{"code", "research", "review", "none"}
	for _, name := range required {
		if !names[name] {
			t.Errorf("missing mode: %s", name)
		}
	}
}

func TestGetModeFound(t *testing.T) {
	InitModes("/tmp/modes")

	m, ok := GetMode("code")
	if !ok {
		t.Fatal("expected to find 'code' mode")
	}
	if m.Name != "code" {
		t.Errorf("Name = %q, want %q", m.Name, "code")
	}
}

func TestGetModeNotFound(t *testing.T) {
	InitModes("/tmp/modes")

	_, ok := GetMode("nonexistent")
	if ok {
		t.Error("expected false for nonexistent mode")
	}
}

func TestGetModeEmptyName(t *testing.T) {
	InitModes("/tmp/modes")

	_, ok := GetMode("")
	if ok {
		t.Error("expected false for empty mode name")
	}
}

func TestGetModeResearchHasPromptFile(t *testing.T) {
	InitModes("/tmp/modes")

	m, ok := GetMode("research")
	if !ok {
		t.Fatal("expected to find 'research' mode")
	}
	if m.Flag != "--append-system-prompt-file" {
		t.Errorf("Flag = %q, want %q", m.Flag, "--append-system-prompt-file")
	}
	if m.Value == "" {
		t.Error("expected Value to be set for research mode")
	}
}

func TestGetModeCodeHasNoPromptFile(t *testing.T) {
	InitModes("/tmp/modes")

	m, ok := GetMode("code")
	if !ok {
		t.Fatal("expected to find 'code' mode")
	}
	if m.Flag != "" {
		t.Errorf("code mode should have no Flag, got %q", m.Flag)
	}
	if m.Value != "" {
		t.Errorf("code mode should have no Value, got %q", m.Value)
	}
}

func TestGetModeNoneHasNoPromptFile(t *testing.T) {
	InitModes("/tmp/modes")

	m, ok := GetMode("none")
	if !ok {
		t.Fatal("expected to find 'none' mode")
	}
	if m.Flag != "" || m.Value != "" {
		t.Error("none mode should have no Flag or Value")
	}
}

func TestInitModesOverwritesPrevious(t *testing.T) {
	InitModes("/path/a")
	first := len(Modes)

	InitModes("/path/b")
	second := len(Modes)

	if first != second {
		t.Errorf("InitModes should overwrite, not append: %d vs %d", first, second)
	}
}
