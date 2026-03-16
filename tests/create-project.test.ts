import { test, expect } from "@microsoft/tui-test";
import { createTestEnv, IARA_BIN, waitForReady } from "./helpers.js";

const env = createTestEnv();

test.use({
  program: { file: IARA_BIN },
  rows: 24,
  columns: 80,
  env: env.env,
});

test.describe("Create Project", () => {
  test("shows name input step", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
    await expect(terminal.getByText(/Project name/g)).toBeVisible();
  });

  test("accepts text input for project name", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
    terminal.write("my-test-project");
    await expect(terminal.getByText("my-test-project")).toBeVisible();
  });

  test("shows method selection after entering name", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
    terminal.write("test-proj");
    terminal.submit();
    await expect(terminal.getByText(/Empty/g)).toBeVisible();
  });

  test("returns to project list on Escape from name step", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
    terminal.keyEscape();
    await expect(terminal.getByText("PROJECTS")).toBeVisible();
  });

  test("returns to name step on Escape from method step", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("n");
    await expect(terminal.getByText("NEW PROJECT")).toBeVisible();
    terminal.write("esc-test");
    terminal.submit();
    await expect(terminal.getByText(/Empty/g)).toBeVisible();
    terminal.keyEscape();
    await expect(terminal.getByText(/Project name/g)).toBeVisible();
  });
});
