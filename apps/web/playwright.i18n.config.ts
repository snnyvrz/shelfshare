import { defineConfig, devices } from "@playwright/test";
import { fileURLToPath } from "node:url";

export default defineConfig({
    testDir: "./e2e",
    testMatch: "localization.spec.ts",
    outputDir: "./test-results/i18n",
    workers: 1,
    timeout: 30000,
    use: { baseURL: "http://localhost:5174", trace: "retain-on-failure" },
    projects: [
        {
            name: "chromium",
            use: { ...devices["Desktop Chrome"], channel: process.env.PLAYWRIGHT_CHANNEL || undefined },
        },
    ],
    webServer: [
        {
            command: "node --import tsx apps/web/tests/fixtures/localization-api.ts",
            cwd: fileURLToPath(new URL("../..", import.meta.url)),
            url: "http://127.0.0.1:31303/health",
        },
        {
            command: "bun x vite dev --host 127.0.0.1 --port 5174 --strictPort",
            cwd: fileURLToPath(new URL(".", import.meta.url)),
            url: "http://localhost:5174",
            env: {
                JWT_SECRET: "localization-test-secret",
                AUTH_API_URL: "http://127.0.0.1:31303/api/auth",
                BOOKS_API_URL: "http://127.0.0.1:31303/api",
            },
        },
    ],
});
