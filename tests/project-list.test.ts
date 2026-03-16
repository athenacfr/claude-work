import { test, expect } from "@microsoft/tui-test";
import { createTestEnv, IARA_BIN, waitForReady } from "./helpers.js";

const emptyEnv = createTestEnv();

test.use({
  program: { file: IARA_BIN },
  rows: 24,
  columns: 80,
  env: emptyEnv.env,
});

test.describe("Project List — empty", () => {
  test("shows PROJECTS header on launch", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText("PROJECTS")).toBeVisible();
  });

  test("shows new project entry even when empty", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText(/new project/g, { strict: false })).toBeVisible();
  });

  test("shows permissions and compact settings", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText(/ompact/g)).toBeVisible();
  });

  test("quits on Ctrl+C", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.keyCtrlC();
    terminal.onExit((exit) => {
      expect(exit.exitCode).toBe(0);
    });
  });
});
