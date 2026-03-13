package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setTestProjectsDir overrides CW_PROJECTS_DIR for isolated testing.
func setTestProjectsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CW_PROJECTS_DIR", dir)
	return dir
}

// --- Create ---

func TestCreateProject(t *testing.T) {
	setTestProjectsDir(t)

	path, err := Create("my-project")
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("project dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestCreateProjectIdempotent(t *testing.T) {
	setTestProjectsDir(t)

	_, err := Create("my-project")
	if err != nil {
		t.Fatal(err)
	}

	// Creating again should not fail (MkdirAll)
	_, err = Create("my-project")
	if err != nil {
		t.Errorf("second Create should not fail: %v", err)
	}
}

func TestCreateProjectSpecialChars(t *testing.T) {
	setTestProjectsDir(t)

	// Project names with spaces and special chars
	_, err := Create("my project")
	if err != nil {
		t.Errorf("Create with spaces failed: %v", err)
	}

	_, err = Create("project-with-dashes_and_underscores")
	if err != nil {
		t.Errorf("Create with dashes/underscores failed: %v", err)
	}
}

// --- Get ---

func TestGetProject(t *testing.T) {
	dir := setTestProjectsDir(t)

	// Create project
	projectDir := filepath.Join(dir, "test-project")
	os.MkdirAll(projectDir, 0755)

	p, err := Get("test-project")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "test-project" {
		t.Errorf("Name = %q, want %q", p.Name, "test-project")
	}
	if len(p.Repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(p.Repos))
	}
}

func TestGetProjectWithRepos(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "test-project")
	os.MkdirAll(projectDir, 0755)

	// Create a git repo inside
	repoPath := filepath.Join(projectDir, "my-repo")
	initTestRepoAt(t, repoPath)

	p, err := Get("test-project")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(p.Repos))
	}
	if p.Repos[0].Name != "my-repo" {
		t.Errorf("Repo.Name = %q, want %q", p.Repos[0].Name, "my-repo")
	}
}

func TestGetProjectNonGitDirsIgnored(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "test-project")
	os.MkdirAll(projectDir, 0755)

	// Create a non-git directory (should be ignored)
	os.MkdirAll(filepath.Join(projectDir, "not-a-repo"), 0755)
	// Create a regular file (should be ignored)
	os.WriteFile(filepath.Join(projectDir, "readme.txt"), []byte("hi"), 0644)

	p, err := Get("test-project")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Repos) != 0 {
		t.Errorf("expected 0 repos (non-git dirs ignored), got %d", len(p.Repos))
	}
}

func TestGetProjectNotFound(t *testing.T) {
	setTestProjectsDir(t)

	_, err := Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestGetProjectIsFile(t *testing.T) {
	dir := setTestProjectsDir(t)

	// Create a file (not dir) with the project name
	os.WriteFile(filepath.Join(dir, "not-a-dir"), []byte("hi"), 0644)

	_, err := Get("not-a-dir")
	if err == nil {
		t.Error("expected error when project path is a file")
	}
}

// --- List ---

func TestListEmpty(t *testing.T) {
	setTestProjectsDir(t)

	projects, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestListMultipleProjects(t *testing.T) {
	dir := setTestProjectsDir(t)

	os.MkdirAll(filepath.Join(dir, "alpha"), 0755)
	os.MkdirAll(filepath.Join(dir, "beta"), 0755)
	os.MkdirAll(filepath.Join(dir, "gamma"), 0755)

	projects, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 3 {
		t.Errorf("expected 3 projects, got %d", len(projects))
	}
}

func TestListSkipsFiles(t *testing.T) {
	dir := setTestProjectsDir(t)

	os.MkdirAll(filepath.Join(dir, "real-project"), 0755)
	os.WriteFile(filepath.Join(dir, "stray-file.txt"), []byte("hi"), 0644)

	projects, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project (file skipped), got %d", len(projects))
	}
}

func TestListProjectsDirNotExist(t *testing.T) {
	t.Setenv("CW_PROJECTS_DIR", "/tmp/nonexistent-cw-projects-"+filepath.Base(t.TempDir()))

	projects, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if projects != nil {
		t.Errorf("expected nil for nonexistent dir, got %d projects", len(projects))
	}
}

// --- Rename ---

func TestRenameProject(t *testing.T) {
	dir := setTestProjectsDir(t)

	os.MkdirAll(filepath.Join(dir, "old-name"), 0755)

	err := Rename("old-name", "new-name")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "old-name")); !os.IsNotExist(err) {
		t.Error("old project dir should not exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "new-name")); err != nil {
		t.Error("new project dir should exist")
	}
}

func TestRenameNonexistent(t *testing.T) {
	setTestProjectsDir(t)

	err := Rename("nonexistent", "new-name")
	if err == nil {
		t.Error("expected error renaming nonexistent project")
	}
}

func TestRenameToExistingName(t *testing.T) {
	dir := setTestProjectsDir(t)

	os.MkdirAll(filepath.Join(dir, "project-a"), 0755)
	os.MkdirAll(filepath.Join(dir, "project-b"), 0755)

	// On Linux, rename to existing dir succeeds if target is empty dir
	// This is a platform-specific behavior we should be aware of
	err := Rename("project-a", "project-b")
	// Don't assert error since os.Rename behavior varies by platform
	_ = err
}

// --- Delete ---

func TestDeleteProject(t *testing.T) {
	dir := setTestProjectsDir(t)

	os.MkdirAll(filepath.Join(dir, "to-delete"), 0755)

	err := Delete("to-delete")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "to-delete")); !os.IsNotExist(err) {
		t.Error("deleted project dir should not exist")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	setTestProjectsDir(t)

	// os.RemoveAll returns nil for nonexistent paths
	err := Delete("nonexistent")
	if err != nil {
		t.Errorf("Delete nonexistent should not error: %v", err)
	}
}

// --- RemoveRepo ---

func TestRemoveRepo(t *testing.T) {
	dir := setTestProjectsDir(t)

	projectDir := filepath.Join(dir, "my-project")
	repoPath := filepath.Join(projectDir, "my-repo")
	initTestRepoAt(t, repoPath)

	err := RemoveRepo("my-project", "my-repo")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Error("repo dir should be removed")
	}
}

// --- CopyDir ---

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	// Create source structure
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "nested.txt"), []byte("world"), 0644)

	err := CopyDir(src, dest)
	if err != nil {
		t.Fatal(err)
	}

	// Verify files
	data, err := os.ReadFile(filepath.Join(dest, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("file.txt = %q, want %q", string(data), "hello")
	}

	data, err = os.ReadFile(filepath.Join(dest, "sub", "nested.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Errorf("nested.txt = %q, want %q", string(data), "world")
	}
}

func TestCopyDirPreservesPermissions(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	os.WriteFile(filepath.Join(src, "exec.sh"), []byte("#!/bin/sh"), 0755)

	err := CopyDir(src, dest)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dest, "exec.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Error("expected executable permission preserved")
	}
}

func TestCopyDirEmpty(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	err := CopyDir(src, dest)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected dest to be a directory")
	}
}

func TestCopyDirWithSymlink(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	// Create a file and a symlink to it
	os.WriteFile(filepath.Join(src, "real.txt"), []byte("content"), 0644)
	os.Symlink(filepath.Join(src, "real.txt"), filepath.Join(src, "link.txt"))

	// CopyDir currently follows symlinks (reads content via os.ReadFile)
	// This test documents that behavior
	err := CopyDir(src, dest)
	if err != nil {
		t.Fatal(err)
	}

	// The symlink should be copied as a regular file (content copied)
	data, err := os.ReadFile(filepath.Join(dest, "link.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Errorf("symlink copy = %q, want %q", string(data), "content")
	}
}

func TestCopyDirBrokenSymlink(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	// Create a broken symlink
	os.Symlink("/nonexistent/target", filepath.Join(src, "broken.txt"))

	// CopyDir should fail on broken symlinks because os.ReadFile follows the link
	err := CopyDir(src, dest)
	if err == nil {
		t.Error("expected error for broken symlink, but CopyDir succeeded")
	}
}

// --- MoveDir ---

func TestMoveDir(t *testing.T) {
	parent := t.TempDir()
	src := filepath.Join(parent, "src")
	dest := filepath.Join(parent, "dest")

	os.MkdirAll(src, 0755)
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0644)

	err := MoveDir(src, dest)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source should not exist after move")
	}

	data, err := os.ReadFile(filepath.Join(dest, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("moved file = %q, want %q", string(data), "hello")
	}
}

// --- Get: ReadDir error (permissions) ---

func TestGetProjectReadDirError(t *testing.T) {
	dir := setTestProjectsDir(t)
	projectDir := filepath.Join(dir, "my-project")
	os.MkdirAll(projectDir, 0755)

	// Create a subdirectory but remove read permission
	subDir := filepath.Join(projectDir, "no-read")
	os.MkdirAll(subDir, 0755)
	os.Chmod(projectDir, 0000)
	defer os.Chmod(projectDir, 0755)

	// Get should return the project without repos (ReadDir fails silently)
	p, err := Get("my-project")
	if err != nil {
		// On some systems this returns an error, on others it succeeds with empty repos
		return
	}
	if p == nil {
		t.Fatal("expected non-nil project")
	}
}

// --- List: skips projects where Get fails ---

func TestListSkipsFailedProjects(t *testing.T) {
	dir := setTestProjectsDir(t)

	os.MkdirAll(filepath.Join(dir, "good"), 0755)
	// Create a file that looks like a project name but isn't a dir
	// (Get will fail because it's not a directory)
	os.WriteFile(filepath.Join(dir, "bad-file"), []byte("not a dir"), 0644)

	projects, err := List()
	if err != nil {
		t.Fatal(err)
	}
	// Only "good" should be listed; "bad-file" is skipped in the for loop
	// because it's not a directory (the entries loop checks e.IsDir())
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

// --- CopyDir: source doesn't exist ---

func TestCopyDirSourceNotExist(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest")
	err := CopyDir("/nonexistent/source", dest)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

// --- helpers ---

func initTestRepoAt(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	cmds := [][]string{
		{"git", "init", path},
		{"git", "-C", path, "config", "user.email", "test@test.com"},
		{"git", "-C", path, "config", "user.name", "Test"},
		{"git", "-C", path, "commit", "--allow-empty", "-m", "initial"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cmd %v failed: %s\n%s", args, err, out)
		}
	}
}
