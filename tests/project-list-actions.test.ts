import { test, expect } from "@microsoft/tui-test";
import { createTestEnv, createFakeProject, IARA_BIN, waitForReady } from "./helpers.js";

const env = createTestEnv();
createFakeProject(env.projectsDir, "action-project", {
  repos: ["main-repo"],
  metadata: { title: "Action Project", description: "For testing actions" },
});

test.use({
  program: { file: IARA_BIN },
  rows: 24,
  columns: 80,
  env: env.env,
});

test.describe("Project List — rename", () => {
  test("enters rename mode with r", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText("action-project", { strict: false })).toBeVisible();
    terminal.write("r");
    await expect(terminal.getByText(/Rename project/g)).toBeVisible();
  });

  test("shows confirm/cancel hints in rename mode", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("r");
    await expect(terminal.getByText(/confirm/g)).toBeVisible();
    await expect(terminal.getByText(/cancel/g)).toBeVisible();
  });

  test("cancels rename with Escape", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("r");
    await expect(terminal.getByText(/Rename project/g)).toBeVisible();
    terminal.keyEscape();
    await expect(terminal.getByText("action-project", { strict: false })).toBeVisible();
  });
});

test.describe("Project List — delete", () => {
  test("shows delete confirmation with d", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("d");
    await expect(terminal.getByText(/Delete project/g)).toBeVisible();
    await expect(terminal.getByText(/confirm/g)).toBeVisible();
  });

  test("cancels delete with n", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("d");
    await expect(terminal.getByText(/Delete project/g)).toBeVisible();
    terminal.write("n");
    await expect(terminal.getByText("action-project", { strict: false })).toBeVisible();
  });
});

test.describe("Project List — navigation to create", () => {
  test("pressing n opens create project screen", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
  });
});
