import { test, expect } from "@microsoft/tui-test";
import { createTestEnv, createFakeProject, CW_BIN, waitForReady } from "./helpers.js";

const env = createTestEnv();
createFakeProject(env.projectsDir, "edit-target", {
  repos: ["service-a", "service-b"],
  metadata: { title: "Edit Target", description: "For edit tests" },
});

test.use({
  program: { file: CW_BIN },
  rows: 24,
  columns: 80,
  env: env.env,
});

test.describe("Edit Project", () => {
  test("expands to show repos", async ({ terminal }) => {
    await waitForReady(terminal);
    await expect(terminal.getByText("edit-target", { strict: false })).toBeVisible();
    terminal.write("t");
    await expect(terminal.getByText("service-a", { strict: false })).toBeVisible();
    await expect(terminal.getByText("service-b", { strict: false })).toBeVisible();
  });

  test("shows add repo option when expanded", async ({ terminal }) => {
    await waitForReady(terminal);
    terminal.write("t");
    await expect(terminal.getByText(/add repo/g)).toBeVisible();
  });
});
