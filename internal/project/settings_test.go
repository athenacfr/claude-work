package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadGlobalSettings(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CW_DATA_DIR", dir)

	s := Settings{BypassPermissions: true, AutoCompactLimit: 60}
	if err := SaveGlobalSettings(s); err != nil {
		t.Fatal(err)
	}

	loaded := LoadGlobalSettings()
	if loaded.BypassPermissions != true {
		t.Error("expected BypassPermissions=true")
	}
	if loaded.AutoCompactLimit != 60 {
		t.Errorf("AutoCompactLimit = %d, want 60", loaded.AutoCompactLimit)
	}
}

func TestLoadGlobalSettingsDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CW_DATA_DIR", dir)

	// No settings file exists — should return default
	s := LoadGlobalSettings()
	if !s.BypassPermissions {
		t.Error("default should have BypassPermissions=true")
	}
}

func TestLoadGlobalSettingsCorrupted(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CW_DATA_DIR", dir)

	os.WriteFile(filepath.Join(dir, "settings.json"), []byte("not json"), 0644)

	s := LoadGlobalSettings()
	if !s.BypassPermissions {
		t.Error("corrupted file should return default with BypassPermissions=true")
	}
}

func TestSaveGlobalSettingsOverwrites(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CW_DATA_DIR", dir)

	SaveGlobalSettings(Settings{BypassPermissions: true})
	SaveGlobalSettings(Settings{BypassPermissions: false, AutoCompactLimit: 40})

	loaded := LoadGlobalSettings()
	if loaded.BypassPermissions != false {
		t.Error("expected BypassPermissions=false after overwrite")
	}
	if loaded.AutoCompactLimit != 40 {
		t.Errorf("AutoCompactLimit = %d, want 40", loaded.AutoCompactLimit)
	}
}

func TestGlobalSettingsPath(t *testing.T) {
	t.Setenv("CW_DATA_DIR", "/test/data")
	got := globalSettingsPath()
	if got != "/test/data/settings.json" {
		t.Errorf("globalSettingsPath = %q, want %q", got, "/test/data/settings.json")
	}
}
