import { test, expect } from "@microsoft/tui-test";
import { createTestEnv, createFakeProject, IARA_BIN, waitForReady } from "./helpers.js";

const env = createTestEnv();
createFakeProject(env.projectsDir, "alpha-project", {
  repos: ["frontend", "backend"],
  metadata: { title: "Alpha", description: "Test project alpha" },
});
createFakeProject(env.projectsDir, "beta-project", {
  repos: ["api"],
  metadata: { title: "Beta", description: "Test project beta" },
});

test.use({
  program: { file: IARA_BIN },
  rows: 24,
  columns: 80,
  env: env.env,
});

test.describe("Project List — navigation", () => {
  test("lists project names", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText("alpha-project", { strict: false })).toBeVisible();
    await expect(terminal.getByText("beta-project", { strict: false })).toBeVisible();
  });

  test("navigates down with j", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("j");
    await expect(terminal.getByText("beta-project", { strict: false })).toBeVisible();
  });

  test("navigates down with arrow key", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.keyDown();
    await expect(terminal.getByText("beta-project", { strict: false })).toBeVisible();
  });

  test("shows + new project entry", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText(/new project/g)).toBeVisible();
  });

  test("expands repos with t", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("t");
    await expect(terminal.getByText("frontend", { strict: false })).toBeVisible();
    await expect(terminal.getByText("backend", { strict: false })).toBeVisible();
  });

  test("expands repos with space", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.keyPress(" ");
    await expect(terminal.getByText("frontend", { strict: false })).toBeVisible();
  });

  test("collapses repos by toggling t", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("t");
    await expect(terminal.getByText("frontend", { strict: false })).toBeVisible();
    terminal.write("t");
    await expect(terminal).toMatchSnapshot();
  });

  test("expands with l and collapses with h", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("l");
    await expect(terminal.getByText("frontend", { strict: false })).toBeVisible();
    terminal.write("h");
    await expect(terminal).toMatchSnapshot();
  });

  test("toggles permissions with p", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText(/bypass/g)).toBeVisible();
    terminal.write("p");
    await expect(terminal.getByText(/normal/g, { strict: false })).toBeVisible();
    terminal.write("p");
    await expect(terminal.getByText(/bypass/g)).toBeVisible();
  });

  test("cycles auto-compact with c", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("c");
    await expect(terminal.getByText(/40%/g)).toBeVisible();
    terminal.write("c");
    await expect(terminal.getByText(/50%/g)).toBeVisible();
  });

  test("enters search mode with s and filters", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("s");
    terminal.write("beta");
    await expect(terminal.getByText("beta-project", { strict: false })).toBeVisible();
    await expect(terminal).toMatchSnapshot();
  });

  test("exits search with Escape restores list", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("s");
    terminal.write("beta");
    await expect(terminal.getByText("beta-project", { strict: false })).toBeVisible();
    terminal.keyEscape();
    await expect(terminal.getByText("alpha-project", { strict: false })).toBeVisible();
    await expect(terminal.getByText("beta-project", { strict: false })).toBeVisible();
  });
});
