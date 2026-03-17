import { mkdtempSync, mkdirSync, writeFileSync, rmSync, utimesSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";
import { test, expect } from "@microsoft/tui-test";

// tui-test compiles to a cache dir, so we resolve the binary path
// relative to the project root using an env var or a known absolute path.
const PROJECT_ROOT = process.env.IARA_TEST_ROOT ?? join(dirname(fileURLToPath(import.meta.url)), "..");

// Resolved path to the iara binary (built by `make build` before tests run)
export const IARA_BIN = join(PROJECT_ROOT, "bin", "iara");

// Create an isolated temp environment for a test suite.
// Returns paths and a cleanup function.
export function createTestEnv() {
  const root = mkdtempSync(join(tmpdir(), "iara-test-"));
  const projectsDir = join(root, "projects");
  const dataDir = join(root, "data");
  mkdirSync(projectsDir, { recursive: true });
  mkdirSync(dataDir, { recursive: true });
  return {
    root,
    projectsDir,
    dataDir,
    env: {
      IARA_PROJECTS_DIR: projectsDir,
      IARA_DATA_DIR: dataDir,
      // Prevent iara from picking up real user config
      HOME: root,
    },
    cleanup() {
      rmSync(root, { recursive: true, force: true });
    },
  };
}

// Create a fake project inside the test env's projects dir.
// Initializes a bare git repo so iara recognizes it.
export function createFakeProject(
  projectsDir: string,
  name: string,
  opts?: { repos?: string[]; metadata?: { title?: string; description?: string; instructions?: string } }
) {
  const projectDir = join(projectsDir, name);
  mkdirSync(projectDir, { recursive: true });

  // Create .iara metadata dir
  const iaraDir = join(projectDir, ".iara");
  mkdirSync(iaraDir, { recursive: true });

  if (opts?.metadata) {
    writeFileSync(
      join(iaraDir, "metadata.json"),
      JSON.stringify(opts.metadata, null, 2)
    );
  }

  // Create repo subdirectories with git init
  const repos = opts?.repos ?? ["repo-one"];
  for (const repo of repos) {
    const repoDir = join(projectDir, repo);
    mkdirSync(repoDir, { recursive: true });
    execSync("git init -q && git commit --allow-empty -m init -q", {
      cwd: repoDir,
      stdio: "ignore",
    });
  }

  return projectDir;
}

// Touch project directories so "opened just now" stays fresh across retries.
// Call this in test.beforeEach() for tests with snapshot assertions that include timestamps.
export function touchProjectDirs(...dirs: string[]) {
  const now = new Date();
  for (const dir of dirs) {
    utimesSync(dir, now, now);
  }
}

// Configure tui-test to launch the iara binary with isolated env.
// Call this at the top of each test file.
export function useIara(env: Record<string, string | undefined>) {
  test.use({
    program: { file: IARA_BIN },
    rows: 24,
    columns: 80,
    env,
  });
}

// Wait for the TUI to be fully rendered by checking for the footer bar.
// The footer (permissions/compact) is the last thing rendered, so its presence
// indicates the full TUI frame has been drawn.
// Optionally touches project directories to keep "opened just now" fresh.
export async function waitForReady(terminal: any, projectDirs?: string[]) {
  if (projectDirs) {
    touchProjectDirs(...projectDirs);
  }
  await expect(terminal.getByText(/permissions/g)).toBeVisible();
}
