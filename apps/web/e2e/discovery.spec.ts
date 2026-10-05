import { test, expect } from "@playwright/test";
import { visit } from "./helpers";

test("city opt-in discovers readers and books by same city and approximate radius", async ({ page, request }) => {
    const stamp = Date.now().toString();
    const authApi = process.env.AUTH_API_URL || "http://localhost:3030/api/auth";
    const booksApi = process.env.BOOKS_API_URL || "http://localhost:8080/api";
    const email = `discovery-${stamp}@example.com`;
    const password = "discovery-e2e-password";
    const name = `Discovery reader ${stamp}`;
    const title = `Discovery story ${stamp}`;
    let token = "",
        copyId = "";
    try {
        await visit(page, "/register");
        await page.getByLabel("Public display name").fill(name);
        await page.getByLabel("Email address").fill(email);
        await page.getByLabel("Password", { exact: true }).fill(password);
        await page.getByRole("button", { name: "Join the shelf", exact: true }).click();
        await expect(page).toHaveURL(/\/books$/);
        const account = await request.post(`${authApi}/login`, { data: { email, password } });
        token = (await account.json()).token;
        const headers = { Authorization: `Bearer ${token}` };
        const author = await request.post(`${booksApi}/authors`, {
            headers,
            data: { name: `Discovery author ${stamp}` },
        });
        expect(author.status()).toBe(201);
        const book = await request.post(`${booksApi}/books`, {
            headers,
            data: { title, author_id: (await author.json()).data.id },
        });
        expect(book.status()).toBe(201);
        const copy = await request.post(`${booksApi}/copies`, {
            headers,
            data: { bookId: (await book.json()).data.id, condition: "Good", visible: true, lendable: true },
        });
        expect(copy.status()).toBe(201);
        copyId = (await copy.json()).id;

        await visit(page, "/account/profile");
        await expect(page.getByLabel("Show me and my public books in nearby discovery")).not.toBeChecked();
        await page.getByLabel("City", { exact: true }).selectOption("112931");
        await page.getByLabel("Show me and my public books in nearby discovery").check();
        await page.getByRole("button", { name: "Save profile" }).click();
        await expect(page.getByRole("status")).toContainText("Profile saved");

        await visit(page, "/shelves");
        await page.getByLabel("Discovery", { exact: true }).selectOption("city");
        await page.getByLabel("Search from city").selectOption("112931");
        await page.getByRole("button", { name: "Search shelves" }).click();
        await expect(page.getByRole("heading", { name, exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: title, exact: true })).toBeVisible();
        await expect(
            page.locator("article").filter({ has: page.getByRole("heading", { name, exact: true }) })
        ).toContainText("Tehran, Iran");

        await page.getByLabel("Discovery", { exact: true }).selectOption("radius");
        await page.getByLabel("Search from city").selectOption("128747");
        await page.getByLabel("Radius (km)").fill("25");
        await page.getByRole("button", { name: "Search shelves" }).click();
        await expect(page.getByRole("heading", { name, exact: true })).toHaveCount(0);
        await expect(page.getByRole("heading", { name: title, exact: true })).toHaveCount(0);
        await page.getByLabel("Radius (km)").fill("50");
        await page.getByRole("button", { name: "Search shelves" }).click();
        await expect(page.getByRole("heading", { name, exact: true })).toBeVisible();
        await expect(page.getByRole("heading", { name: title, exact: true })).toBeVisible();

        await visit(page, "/account/profile");
        await page.getByLabel("Show me and my public books in nearby discovery").uncheck();
        await page.getByRole("button", { name: "Save profile" }).click();
        await expect(page.getByRole("status")).toContainText("Profile saved");
        await visit(page, "/shelves?mode=city&cityId=112931");
        await expect(page.getByRole("heading", { name, exact: true })).toHaveCount(0);
        await expect(page.getByRole("heading", { name: title, exact: true })).toHaveCount(0);
    } finally {
        if (token) {
            const headers = { Authorization: `Bearer ${token}` };
            await request.patch(`${authApi}/me`, { headers, data: { discoveryEnabled: false } });
            if (copyId) await request.delete(`${booksApi}/copies/${copyId}`, { headers });
        }
    }
});
