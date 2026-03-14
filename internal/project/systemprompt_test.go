package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSystemPromptSingleRepo(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	initTestRepoAt(t, filepath.Join(projectDir, "my-repo"))

	prompt, err := BuildSystemPrompt("my-project")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(prompt, "single subproject") {
		t.Error("expected single subproject template")
	}
	if !strings.Contains(prompt, "`my-repo/`") {
		t.Error("expected subproject name in prompt")
	}
}

func TestBuildSystemPromptMultiRepo(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	initTestRepoAt(t, filepath.Join(projectDir, "frontend"))
	initTestRepoAt(t, filepath.Join(projectDir, "backend"))

	prompt, err := BuildSystemPrompt("my-project")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(prompt, "multiple subprojects") {
		t.Error("expected multi subproject template")
	}
	if !strings.Contains(prompt, "`frontend/`") {
		t.Error("expected frontend in prompt")
	}
	if !strings.Contains(prompt, "`backend/`") {
		t.Error("expected backend in prompt")
	}
}

func TestBuildSystemPromptNoRepos(t *testing.T) {
	dir := setTestProjectsDir(t)
	os.MkdirAll(filepath.Join(dir, "empty-project"), 0755)

	prompt, err := BuildSystemPrompt("empty-project")
	if err != nil {
		t.Fatal(err)
	}

	// No subprojects: uses multi template with empty list
	if prompt == "" {
		t.Error("expected non-empty prompt even with no repos")
	}
}

func TestBuildSystemPromptNonexistent(t *testing.T) {
	setTestProjectsDir(t)

	_, err := BuildSystemPrompt("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestBuildSystemPromptContainsRules(t *testing.T) {
	dir := setTestProjectsDir(t)
	initTestRepoAt(t, filepath.Join(dir, "my-project", "repo"))

	prompt, err := BuildSystemPrompt("my-project")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(prompt, "Rules") {
		t.Error("expected Rules section in prompt")
	}
	if !strings.Contains(prompt, "Do NOT create code files") {
		t.Error("expected file creation rule")
	}
}
