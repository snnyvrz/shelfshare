import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ command }) => {
    if (command === "serve") {
        const local = loadEnv("dev", "../..", "");
        for (const key of ["JWT_SECRET", "BOOKS_API_URL", "AUTH_API_URL", "BOOKS_WS_URL"]) {
            if (!process.env[key] && local[key]) process.env[key] = local[key];
        }
    }
    return { plugins: [sveltekit()], envDir: "../.." };
});
