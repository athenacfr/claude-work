package project

import (
	"os"
	"path/filepath"
	"testing"
)

// --- SaveMetadata ---

func TestSaveMetadataValid(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"My Project","description":"A test project","instructions":"Do things"}`

	err := SaveMetadata(dir, raw)
	if err != nil {
		t.Fatal(err)
	}

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(dir, ".iara", "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("metadata.json should not be empty")
	}
}

func TestSaveMetadataInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	err := SaveMetadata(dir, "not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveMetadataMissingTitle(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"","description":"desc","instructions":"inst"}`

	err := SaveMetadata(dir, raw)
	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestSaveMetadataMissingDescription(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"Title","description":"","instructions":"inst"}`

	err := SaveMetadata(dir, raw)
	if err == nil {
		t.Error("expected error for missing description")
	}
}

func TestSaveMetadataMissingInstructions(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"Title","description":"desc","instructions":""}`

	err := SaveMetadata(dir, raw)
	if err == nil {
		t.Error("expected error for missing instructions")
	}
}

func TestSaveMetadataExtraFields(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"Title","description":"desc","instructions":"inst","extra":"field"}`

	// Extra fields should be silently ignored (json.Unmarshal behavior)
	err := SaveMetadata(dir, raw)
	if err != nil {
		t.Errorf("extra fields should not cause error: %v", err)
	}
}

func TestSaveMetadataCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"Title","description":"desc","instructions":"inst"}`

	err := SaveMetadata(dir, raw)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, ".iara"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected .iara to be a directory")
	}
}

func TestSaveMetadataOverwrite(t *testing.T) {
	dir := t.TempDir()

	raw1 := `{"title":"First","description":"desc1","instructions":"inst1"}`
	if err := SaveMetadata(dir, raw1); err != nil {
		t.Fatal(err)
	}

	raw2 := `{"title":"Second","description":"desc2","instructions":"inst2"}`
	if err := SaveMetadata(dir, raw2); err != nil {
		t.Fatal(err)
	}

	m, err := LoadMetadataAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Second" {
		t.Errorf("Title = %q, want %q (overwrite failed)", m.Title, "Second")
	}
}

// --- LoadMetadataAt ---

func TestLoadMetadataAt(t *testing.T) {
	dir := t.TempDir()

	raw := `{"title":"Test","description":"Desc","instructions":"Inst"}`
	if err := SaveMetadata(dir, raw); err != nil {
		t.Fatal(err)
	}

	m, err := LoadMetadataAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Test" {
		t.Errorf("Title = %q, want %q", m.Title, "Test")
	}
	if m.Description != "Desc" {
		t.Errorf("Description = %q, want %q", m.Description, "Desc")
	}
	if m.Instructions != "Inst" {
		t.Errorf("Instructions = %q, want %q", m.Instructions, "Inst")
	}
}

func TestLoadMetadataAtNotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadMetadataAt(dir)
	if err == nil {
		t.Error("expected error for missing metadata")
	}
}

func TestLoadMetadataAtCorrupted(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".iara"), 0755)
	os.WriteFile(filepath.Join(dir, ".iara", "metadata.json"), []byte("not json"), 0644)

	_, err := LoadMetadataAt(dir)
	if err == nil {
		t.Error("expected error for corrupted metadata")
	}
}

// --- HasMetadata ---

func TestHasMetadata(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	os.MkdirAll(projectDir, 0755)

	if HasMetadata("my-project") {
		t.Error("expected false before saving metadata")
	}

	raw := `{"title":"Title","description":"desc","instructions":"inst"}`
	SaveMetadata(projectDir, raw)

	if !HasMetadata("my-project") {
		t.Error("expected true after saving metadata")
	}
}

// --- Edge cases ---

func TestSaveMetadataUnicode(t *testing.T) {
	dir := t.TempDir()
	raw := `{"title":"项目标题","description":"描述说明","instructions":"操作指南"}`

	err := SaveMetadata(dir, raw)
	if err != nil {
		t.Fatal(err)
	}

	m, err := LoadMetadataAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "项目标题" {
		t.Errorf("Title = %q, want Unicode title", m.Title)
	}
}

// --- LoadMetadata (by name, uses ProjectsDir) ---

func TestLoadMetadataByName(t *testing.T) {
	dir := setTestProjectsDir(t)
	projectDir := filepath.Join(dir, "my-project")
	os.MkdirAll(projectDir, 0755)

	raw := `{"title":"Test","description":"Desc","instructions":"Inst"}`
	SaveMetadata(projectDir, raw)

	m, err := LoadMetadata("my-project")
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Test" {
		t.Errorf("Title = %q, want %q", m.Title, "Test")
	}
}

func TestLoadMetadataByNameNotFound(t *testing.T) {
	setTestProjectsDir(t)
	_, err := LoadMetadata("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestSaveMetadataLargeContent(t *testing.T) {
	dir := t.TempDir()

	// Large instructions field
	largeInst := make([]byte, 10000)
	for i := range largeInst {
		largeInst[i] = 'a'
	}
	raw := `{"title":"Title","description":"desc","instructions":"` + string(largeInst) + `"}`

	err := SaveMetadata(dir, raw)
	if err != nil {
		t.Fatal(err)
	}
}
