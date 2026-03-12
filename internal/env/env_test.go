package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMerge(t *testing.T) {
	dir := t.TempDir()

	globalFile := filepath.Join(dir, ".env.test.global")
	overrideFile := filepath.Join(dir, ".env.test.override")

	os.WriteFile(globalFile, []byte("DB_HOST=localhost\nDB_PORT=5432\nAPI_KEY=global-key\n"), 0644)
	os.WriteFile(overrideFile, []byte("API_KEY=override-key\nDEBUG=true\n"), 0644)

	result, err := merge(globalFile, overrideFile)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Override should win for API_KEY
	if got := findLine(result, "API_KEY"); got != "API_KEY=override-key" {
		t.Errorf("expected API_KEY=override-key, got %s", got)
	}

	// Global values should be preserved
	if got := findLine(result, "DB_HOST"); got != "DB_HOST=localhost" {
		t.Errorf("expected DB_HOST=localhost, got %s", got)
	}

	// Override-only values should appear
	if got := findLine(result, "DEBUG"); got != "DEBUG=true" {
		t.Errorf("expected DEBUG=true, got %s", got)
	}
}

func TestMergeSkipsComments(t *testing.T) {
	dir := t.TempDir()

	globalFile := filepath.Join(dir, "global")
	os.WriteFile(globalFile, []byte("# comment\nFOO=bar\n\n  \nexport BAZ=qux\n"), 0644)

	overrideFile := filepath.Join(dir, "override")
	os.WriteFile(overrideFile, []byte(""), 0644)

	result, err := merge(globalFile, overrideFile)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	if got := findLine(result, "FOO"); got != "FOO=bar" {
		t.Errorf("expected FOO=bar, got %s", got)
	}
	if got := findLine(result, "BAZ"); got != "BAZ=qux" {
		t.Errorf("expected BAZ=qux, got %s", got)
	}
}

func TestMergeMissingFiles(t *testing.T) {
	dir := t.TempDir()

	result, err := merge(filepath.Join(dir, "missing1"), filepath.Join(dir, "missing2"))
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if findLine(result, "=") != "" {
		t.Errorf("expected no env vars, got %q", result)
	}
}

func TestSync(t *testing.T) {
	dir := t.TempDir()

	// Create a fake repo dir
	repoDir := filepath.Join(dir, "my-repo")
	os.MkdirAll(repoDir, 0755)

	// Create global env
	envsDir := EnvsDir()
	os.MkdirAll(envsDir, 0755)
	os.WriteFile(filepath.Join(envsDir, ".env.my-repo.global"), []byte("GLOBAL=yes\n"), 0644)

	err := Sync(dir, []string{"my-repo"})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Check generated .env
	data, err := os.ReadFile(filepath.Join(repoDir, ".env"))
	if err != nil {
		t.Fatalf("read .env failed: %v", err)
	}

	if got := findLine(string(data), "GLOBAL"); got != "GLOBAL=yes" {
		t.Errorf("expected GLOBAL=yes, got %s", got)
	}

	// Check override file was created
	overridePath := filepath.Join(dir, ".env.my-repo.override")
	if _, err := os.Stat(overridePath); os.IsNotExist(err) {
		t.Error("expected .env.my-repo.override to be created")
	}

	// Check symlink for global env file was created in project dir
	symlinkPath := filepath.Join(dir, ".env.my-repo.global")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", symlinkPath, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink", symlinkPath)
	}
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("readlink failed: %v", err)
	}
	if target != GlobalPath("my-repo") {
		t.Errorf("symlink target = %s, want %s", target, GlobalPath("my-repo"))
	}
}

func TestSyncPreservesContentWhenSymlinkReplacedByFile(t *testing.T) {
	dir := t.TempDir()

	repoDir := filepath.Join(dir, "my-repo")
	os.MkdirAll(repoDir, 0755)

	// Create global env with initial content
	envsDir := EnvsDir()
	os.MkdirAll(envsDir, 0755)
	globalPath := filepath.Join(envsDir, ".env.my-repo.global")
	os.WriteFile(globalPath, []byte("OLD_VAR=old\n"), 0644)

	// First sync to create the symlink
	if err := Sync(dir, []string{"my-repo"}); err != nil {
		t.Fatalf("initial Sync failed: %v", err)
	}

	// Simulate editor save-by-rename: replace the symlink with a regular file
	symlinkPath := filepath.Join(dir, ".env.my-repo.global")
	os.Remove(symlinkPath)
	os.WriteFile(symlinkPath, []byte("NEW_VAR=new\nUPDATED=true\n"), 0644)

	// Sync again — should preserve the regular file's content into the global file
	if err := Sync(dir, []string{"my-repo"}); err != nil {
		t.Fatalf("second Sync failed: %v", err)
	}

	// The actual global file should now have the new content
	data, err := os.ReadFile(globalPath)
	if err != nil {
		t.Fatalf("read global file: %v", err)
	}
	if got := findLine(string(data), "NEW_VAR"); got != "NEW_VAR=new" {
		t.Errorf("global file: expected NEW_VAR=new, got %q (content: %q)", got, string(data))
	}
	if got := findLine(string(data), "UPDATED"); got != "UPDATED=true" {
		t.Errorf("global file: expected UPDATED=true, got %q", got)
	}

	// The symlink should be restored
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", symlinkPath, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected %s to be a symlink after fix", symlinkPath)
	}

	// The merged .env should include the new content
	envData, err := os.ReadFile(filepath.Join(repoDir, ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	if got := findLine(string(envData), "NEW_VAR"); got != "NEW_VAR=new" {
		t.Errorf(".env: expected NEW_VAR=new, got %q", got)
	}
}

func findLine(content, key string) string {
	for _, line := range splitLines(content) {
		if len(line) > len(key) && line[:len(key)+1] == key+"=" {
			return line
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
