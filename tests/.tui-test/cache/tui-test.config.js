//# hash=2640692c3769e7b77f7bfcd37780809e
//# sourceMappingURL=tui-test.config.js.map

import { defineConfig } from "@microsoft/tui-test";
export default defineConfig({
    timeout: 30000,
    expect: {
        timeout: 5000
    },
    retries: 2,
    trace: true,
    traceFolder: "tui-traces",
    workers: 2
});
