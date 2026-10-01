import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import { defineConfig } from "eslint/config";
import svelte from "eslint-plugin-svelte";

export default defineConfig([
    { ignores: ["**/.svelte-kit/**", "**/build/**", "**/dist/**", "**/test-results/**"] },
    {
        files: ["**/*.{js,mjs,cjs,ts,mts,cts}"],
        plugins: { js },
        extends: ["js/recommended"],
        languageOptions: { globals: { ...globals.browser, ...globals.node } },
    },
    tseslint.configs.recommended,
    ...svelte.configs.recommended,
    {
        files: ["**/*.svelte"],
        languageOptions: { parserOptions: { parser: tseslint.parser } },
        // The web application is deployed at the origin root (kit.paths.base is empty).
        rules: { "svelte/no-navigation-without-resolve": "off" },
    },
]);
