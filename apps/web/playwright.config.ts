import { defineConfig, devices } from "@playwright/test";
import { fileURLToPath } from "node:url";

export default defineConfig({
    testDir: "./e2e",
    testIgnore: "localization.spec.ts",
    outputDir: "./test-results",
    timeout: 60000,
    workers: 1,
    use: { baseURL: "http://localhost:5173", trace: "retain-on-failure" },
    projects: [
        {
            name: "chromium",
            use: { ...devices["Desktop Chrome"], channel: process.env.PLAYWRIGHT_CHANNEL || undefined },
        },
    ],
    webServer: {
        command: "bun x nx serve web",
        cwd: fileURLToPath(new URL("../..", import.meta.url)),
        url: "http://localhost:5173",
        reuseExistingServer: !process.env.CI,
        timeout: 60000,
    },
});
