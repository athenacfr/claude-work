package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncCommandsSingleRepo(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	repoPath := filepath.Join(projectDir, "my-repo")
	initTestRepoAt(t, repoPath)

	// Create a command in the repo
	cmdDir := filepath.Join(repoPath, ".claude", "commands")
	os.MkdirAll(cmdDir, 0755)
	os.WriteFile(filepath.Join(cmdDir, "deploy.md"), []byte("# Deploy\nDeploy the app"), 0644)

	err := SyncCommands("my-project")
	if err != nil {
		t.Fatal(err)
	}

	// For single repo, command name is preserved
	destPath := filepath.Join(projectDir, ".claude", "commands", "deploy.md")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("synced command not found: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, managedMarker) {
		t.Error("synced command should contain managed marker")
	}
	if !strings.Contains(content, "Deploy the app") {
		t.Error("synced command should contain original content")
	}
	if !strings.Contains(content, "my-repo") {
		t.Error("synced command should contain repo name in header")
	}
}

func TestSyncCommandsMultiRepo(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	repo1 := filepath.Join(projectDir, "frontend")
	repo2 := filepath.Join(projectDir, "backend")
	initTestRepoAt(t, repo1)
	initTestRepoAt(t, repo2)

	// Create commands in both repos
	for _, repo := range []string{repo1, repo2} {
		cmdDir := filepath.Join(repo, ".claude", "commands")
		os.MkdirAll(cmdDir, 0755)
		os.WriteFile(filepath.Join(cmdDir, "deploy.md"), []byte("# Deploy"), 0644)
	}

	err := SyncCommands("my-project")
	if err != nil {
		t.Fatal(err)
	}

	// For multi-repo, commands are prefixed with repo name
	destDir := filepath.Join(projectDir, ".claude", "commands")
	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatal(err)
	}

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}

	if !names["frontend:deploy.md"] {
		t.Error("expected frontend:deploy.md")
	}
	if !names["backend:deploy.md"] {
		t.Error("expected backend:deploy.md")
	}
}

func TestSyncCommandsCleansOldFiles(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	repoPath := filepath.Join(projectDir, "my-repo")
	initTestRepoAt(t, repoPath)

	// Create initial command
	cmdDir := filepath.Join(repoPath, ".claude", "commands")
	os.MkdirAll(cmdDir, 0755)
	os.WriteFile(filepath.Join(cmdDir, "old-cmd.md"), []byte("# Old"), 0644)

	SyncCommands("my-project")

	// Now remove the source command and add a new one
	os.Remove(filepath.Join(cmdDir, "old-cmd.md"))
	os.WriteFile(filepath.Join(cmdDir, "new-cmd.md"), []byte("# New"), 0644)

	SyncCommands("my-project")

	destDir := filepath.Join(projectDir, ".claude", "commands")

	// Old command should be removed
	if _, err := os.Stat(filepath.Join(destDir, "old-cmd.md")); !os.IsNotExist(err) {
		t.Error("old synced command should be removed")
	}

	// New command should exist
	if _, err := os.Stat(filepath.Join(destDir, "new-cmd.md")); err != nil {
		t.Error("new synced command should exist")
	}
}

func TestSyncCommandsPreservesUserFiles(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	repoPath := filepath.Join(projectDir, "my-repo")
	initTestRepoAt(t, repoPath)

	// Create a user command (no managed marker)
	destDir := filepath.Join(projectDir, ".claude", "commands")
	os.MkdirAll(destDir, 0755)
	os.WriteFile(filepath.Join(destDir, "user-cmd.md"), []byte("# My custom command"), 0644)

	// Create repo command
	cmdDir := filepath.Join(repoPath, ".claude", "commands")
	os.MkdirAll(cmdDir, 0755)
	os.WriteFile(filepath.Join(cmdDir, "deploy.md"), []byte("# Deploy"), 0644)

	SyncCommands("my-project")

	// User command should still exist
	data, err := os.ReadFile(filepath.Join(destDir, "user-cmd.md"))
	if err != nil {
		t.Fatal("user command should be preserved")
	}
	if string(data) != "# My custom command" {
		t.Error("user command content was modified")
	}
}

func TestSyncCommandsNoRepoCommands(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	repoPath := filepath.Join(projectDir, "my-repo")
	initTestRepoAt(t, repoPath)

	// No .claude/commands/ in repo
	err := SyncCommands("my-project")
	if err != nil {
		t.Errorf("SyncCommands should not error when repos have no commands: %v", err)
	}
}

func TestSyncCommandsNonexistentProject(t *testing.T) {
	setTestProjectsDir(t)

	err := SyncCommands("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}
