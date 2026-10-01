import { test, expect } from "@playwright/test";

test("a reader can browse, join, contribute, filter, paginate, and log out", async ({ page, request }, testInfo) => {
    const stamp = `${Date.now()}`;
    const email = `shelfshare-e2e-${stamp}@example.com`;
    const password = "test-reader-password";
    const authorName = `E2E Author ${stamp}`;
    const title = `E2E Story ${stamp}`;
    const booksApi = process.env.BOOKS_API_URL || "http://localhost:8080/api";
    const authApi = process.env.AUTH_API_URL || "http://localhost:3030/api/auth";
    const ids: string[] = [];
    let authorId = "";
    let token = "";
    const browserErrors: string[] = [];
    page.on("pageerror", (error) => browserErrors.push(error.message));

    await page.goto("/");
    await expect(page.getByRole("heading", { level: 1 })).toContainText("A shared shelf.");
    await page.screenshot({ path: testInfo.outputPath("home-desktop.png"), fullPage: true });
    await page.setViewportSize({ width: 390, height: 844 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.screenshot({ path: testInfo.outputPath("home-mobile.png"), fullPage: true });
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto("/books/new");
    await expect(page).toHaveURL(/\/login\?returnTo=/);
    expect((await request.post(`${booksApi}/books`, { data: {} })).status()).toBe(401);
    expect((await request.patch(`${booksApi}/authors/not-an-id`, { data: {} })).status()).toBe(401);

    try {
        await page.goto("/register");
        await page.getByLabel("Email address").fill(email);
        await page.getByLabel("Password", { exact: true }).fill(password);
        await page.getByRole("button", { name: "Join the shelf" }).click();
        await expect(page).toHaveURL(/\/books$/);
        const login = await request.post(`${authApi}/login`, { data: { email, password } });
        expect(login.ok()).toBe(true);
        token = (await login.json()).token;
        const session = (await page.context().cookies()).find((cookie) => cookie.name === "shelfshare_session");
        expect(session?.httpOnly).toBe(true);
        expect(session?.sameSite).toBe("Lax");

        await page.goto("/authors/new");
        await page.getByLabel("Author name").fill(authorName);
        await page.getByLabel("Biography").fill("A writer for the automated reader journey.");
        await page.getByRole("button", { name: "Add author", exact: true }).click();
        await expect(page).toHaveURL(/\/authors\/[0-9a-f-]{36}$/);
        await expect(page.getByRole("heading", { level: 1 })).toHaveText(authorName);
        authorId = page.url().split("/").pop()!;
        await page.getByRole("link", { name: "Edit author" }).click();
        await page.getByLabel("Biography").fill("An updated biography.");
        await page.getByRole("button", { name: "Save changes" }).click();
        await expect(page.getByText("An updated biography.", { exact: true })).toBeVisible();

        await page.goto("/books/new");
        await page.getByLabel("Book title").fill(title);
        await page.getByLabel("Author *", { exact: true }).selectOption(authorId);
        await page.getByLabel("Publication date").fill("2020-06-15");
        await page.getByLabel("Description").fill("A shared journey through a growing library.");
        await page.getByRole("button", { name: "Add book", exact: true }).click();
        await expect(page).toHaveURL(/\/books\/[0-9a-f-]{36}$/);
        await expect(page.getByRole("heading", { level: 1 })).toHaveText(title);
        const bookId = page.url().split("/").pop()!;
        ids.push(bookId);
        await page.getByRole("link", { name: "Edit book" }).click();
        await page.getByLabel("Publication date").fill("");
        await page.getByLabel("Description").fill("Updated story description.");
        await page.getByRole("button", { name: "Save changes" }).click();
        await expect(page.getByText("Publication date not listed", { exact: true })).toBeVisible();
        await expect(page.getByText("Updated story description.", { exact: true })).toBeVisible();

        for (let i = 0; i < 13; i++) {
            const created = await request.post(`${booksApi}/books`, {
                headers: { Authorization: `Bearer ${token}` },
                data: {
                    title: `${title} Volume ${String(i).padStart(2, "0")}`,
                    author_id: authorId,
                    published_at: "2021-01-01",
                },
            });
            expect(created.ok()).toBe(true);
            ids.push((await created.json()).data.id);
        }
        await page.goto(`/books?author_id=${authorId}&sort=title_asc`);
        await expect(page.locator(".book-card")).toHaveCount(12);
        await page.getByRole("link", { name: "Next →", exact: true }).click();
        await expect(page).toHaveURL(/page=2/);
        await expect(page.locator(".book-card")).toHaveCount(2);
        await page.goto(`/books?author_id=${authorId}&published_after=2021-01-01&published_before=2021-12-31`);
        await expect(page.getByText("13 books on this shelf")).toBeVisible();
        await page.getByLabel("Search the shelf").fill(title);
        await page.getByRole("button", { name: "Search", exact: true }).click();
        await expect(page).toHaveURL(/q=E2E/);
        await page.setViewportSize({ width: 390, height: 844 });
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
        await page.screenshot({ path: testInfo.outputPath("catalog-mobile.png"), fullPage: true });

        await page.goto(`/books/${bookId}`);
        await page.getByRole("button", { name: "Delete book", exact: true }).click();
        await page.getByRole("button", { name: "Yes, delete", exact: true }).click();
        await expect(page).toHaveURL(/\/books$/);
        await page.getByRole("button", { name: "Log out" }).click();
        await expect(page.getByRole("link", { name: "Log in", exact: true })).toBeVisible();
        await page.goto("/login");
        await page.getByLabel("Email address").fill(email);
        await page.getByLabel("Password", { exact: true }).fill("wrong-password");
        await page.getByRole("button", { name: "Log in", exact: true }).click();
        await expect(page.getByRole("alert")).toContainText("Invalid credentials");
        await page.getByLabel("Password", { exact: true }).fill(password);
        await page.getByRole("button", { name: "Log in", exact: true }).click();
        await expect(page).toHaveURL(/\/books$/);
        for (const id of ids)
            await request.delete(`${booksApi}/books/${id}`, { headers: { Authorization: `Bearer ${token}` } });
        await page.goto(`/authors/${authorId}`);
        await page.getByRole("button", { name: "Delete author", exact: true }).click();
        await page.getByRole("button", { name: "Yes, delete", exact: true }).click();
        await expect(page).toHaveURL(/\/authors$/);
        expect(browserErrors).toEqual([]);
    } catch (cause) {
        console.error("Browser errors:", browserErrors);
        throw cause;
    } finally {
        if (token) {
            for (const id of ids)
                await request.delete(`${booksApi}/books/${id}`, { headers: { Authorization: `Bearer ${token}` } });
            if (authorId)
                await request.delete(`${booksApi}/authors/${authorId}`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
        }
    }
});
