import { test, expect } from "@microsoft/tui-test";
import { createTestEnv, createFakeProject, IARA_BIN, waitForReady } from "./helpers.js";

const env = createTestEnv();
const snapProjectDir = createFakeProject(env.projectsDir, "snap-project", {
  repos: ["web", "api"],
  metadata: { title: "Snap Project", description: "For snapshot tests" },
});
const snapDirs = [snapProjectDir];

test.use({
  program: { file: IARA_BIN },
  rows: 24,
  columns: 80,
  env: env.env,
});

test.describe("Snapshots", () => {
  test("project list initial render", async ({ terminal }) => {
    await waitForReady(terminal, snapDirs);
    await expect(terminal.getByText("snap-project", { strict: false })).toBeVisible();
    await expect(terminal.getByText("Snap Project", { strict: false })).toBeVisible();
    await expect(terminal.getByText("For snapshot tests", { strict: false })).toBeVisible();
    await expect(terminal.getByText(/api on detached/g)).toBeVisible();
    await expect(terminal.getByText(/web on detached/g)).toBeVisible();
    await expect(terminal.getByText(/permissions/g)).toBeVisible();
  });

  test("project list with expanded repos", async ({ terminal }) => {
    await waitForReady(terminal, snapDirs);
    terminal.write("t");
    await expect(terminal.getByText("web", { strict: false })).toBeVisible();
    await expect(terminal).toMatchSnapshot();
  });

  test("project list search active", async ({ terminal }) => {
    await waitForReady(terminal, snapDirs);
    terminal.write("s");
    terminal.write("snap");
    await expect(terminal.getByText("snap-project", { strict: false })).toBeVisible();
  });

  test("mode select initial render", async ({ terminal }) => {
    await waitForReady(terminal, snapDirs);
    terminal.submit();
    // Navigate through task select screen - select default branch
    await expect(terminal.getByText("TASKS")).toBeVisible();
    terminal.keyDown();
    terminal.submit();
    await expect(terminal.getByText("MODE")).toBeVisible();
    await expect(terminal).toMatchSnapshot();
  });

  test("create project name step", async ({ terminal }) => {
    await waitForReady(terminal, snapDirs);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
    await expect(terminal).toMatchSnapshot();
  });

  test("delete confirmation dialog", async ({ terminal }) => {
    await waitForReady(terminal, snapDirs);
    terminal.write("d");
    await expect(terminal.getByText(/Delete project/g)).toBeVisible();
    await expect(terminal).toMatchSnapshot();
  });
});
