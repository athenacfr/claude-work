import { defineConfig } from "@microsoft/tui-test";

export default defineConfig({
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  retries: 2,
  trace: true,
  traceFolder: "tui-traces",
  workers: 2,
});
