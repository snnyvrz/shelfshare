import { expect, type Page } from "@playwright/test";

// SSR forms are visible before hydration. Wait for the application to attach
// bindings before typing, so hydration cannot overwrite automated input.
export async function visit(page: Page, path: string) {
    // Existing domain journeys intentionally exercise the English interface.
    await page.context().addCookies([{ name: "shelfshare_locale", value: "en", url: "http://localhost:5173" }]);
    await page.goto(path);
    await expect(page.locator("html")).toHaveAttribute("data-hydrated", "true");
}
