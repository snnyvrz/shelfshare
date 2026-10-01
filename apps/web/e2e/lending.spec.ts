import { test, expect } from "@playwright/test";
import { visit } from "./helpers";

test("two readers message in real time and lend a physical book through confirmed return", async ({
    browser,
    request,
}, testInfo) => {
    const stamp = Date.now().toString();
    const booksApi = process.env.BOOKS_API_URL || "http://localhost:8080/api";
    const authApi = process.env.AUTH_API_URL || "http://localhost:3030/api/auth";
    const password = "lending-e2e-password";
    const ownerEmail = `owner-${stamp}@example.com`;
    const borrowerEmail = `borrower-${stamp}@example.com`;
    const ownerContext = await browser.newContext();
    const borrowerContext = await browser.newContext();
    const owner = await ownerContext.newPage();
    const borrower = await borrowerContext.newPage();
    const errors: string[] = [];
    owner.on("pageerror", (error) => errors.push(error.message));
    borrower.on("pageerror", (error) => errors.push(error.message));
    let copyId = "";
    let ownerToken = "";
    try {
        for (const [page, email, name] of [
            [owner, ownerEmail, `Owner ${stamp}`],
            [borrower, borrowerEmail, `Borrower ${stamp}`],
        ] as const) {
            await visit(page, "/register");
            await page.getByLabel("Public display name").fill(name);
            await page.getByLabel("Email address").fill(email);
            await page.getByLabel("Password", { exact: true }).fill(password);
            await page.getByRole("button", { name: "Join the shelf", exact: true }).click();
            await expect(page).toHaveURL(/\/books$/);
        }
        const auth = await request.post(`${authApi}/login`, { data: { email: ownerEmail, password } });
        const account = await auth.json();
        ownerToken = account.token;
        const headers = { Authorization: `Bearer ${ownerToken}` };
        const author = await request.post(`${booksApi}/authors`, {
            headers,
            data: { name: `Lending author ${stamp}` },
        });
        expect(author.status()).toBe(201);
        const authorId = (await author.json()).data.id;
        const book = await request.post(`${booksApi}/books`, {
            headers,
            data: { title: `Physical story ${stamp}`, author_id: authorId },
        });
        expect(book.status()).toBe(201);
        const bookId = (await book.json()).data.id;

        await visit(owner, `/my-shelf?book=${bookId}`);
        await owner.getByLabel("Condition", { exact: true }).fill("Good, paperback");
        await owner.getByLabel("Copy notes").fill("English edition");
        await owner.getByRole("button", { name: "Add my copy", exact: true }).click();
        await expect(owner.getByRole("status").filter({ hasText: "Shelf updated" })).toBeVisible();
        const copies = await request.get(`${booksApi}/me/copies`, { headers });
        copyId = (await copies.json()).data.find((copy: { bookId: string }) => copy.bookId === bookId).id;

        const publicReaders = await request.get(`${authApi}/profiles?q=${encodeURIComponent(`Owner ${stamp}`)}`);
        expect(publicReaders.status()).toBe(200);
        const directory = (await publicReaders.json()).data;
        expect(directory.some((profile: { id: string }) => profile.id === account.user.id)).toBe(true);
        expect(
            directory.every((profile: Record<string, unknown>) => !("email" in profile) && !("passwordHash" in profile))
        ).toBe(true);
        await visit(borrower, `/shelves?q=${encodeURIComponent(`Owner ${stamp}`)}`);
        await borrower
            .getByRole("heading", { name: `Owner ${stamp}`, exact: true })
            .getByRole("link")
            .click();
        await expect(borrower.getByRole("heading", { level: 1 })).toContainText(`Owner ${stamp}`);
        await borrower.getByRole("button", { name: `Message Owner ${stamp}` }).click();
        await expect(borrower.getByRole("status").filter({ hasText: "Connected" })).toBeVisible();
        await borrower.getByLabel("Your message").fill("Hello! Could we meet near the library?");
        await borrower.getByRole("button", { name: "Send message", exact: true }).click();
        await expect(borrower.getByRole("log")).toContainText("Hello! Could we meet near the library?");
        const directURL = borrower.url();
        await visit(owner, directURL);
        await expect(owner.getByRole("log")).toContainText("Hello! Could we meet near the library?");
        await owner.getByRole("button", { name: "Accept", exact: true }).click();
        await expect(owner.getByLabel("Your message")).toBeVisible();
        await owner.getByLabel("Your message").fill("Yes, tomorrow at noon works.");
        await owner.getByRole("button", { name: "Send message", exact: true }).click();
        // No navigation or manual refresh: the second browser receives the WebSocket event.
        await expect(borrower.getByRole("log")).toContainText("Yes, tomorrow at noon works.");
        await expect(borrower.getByText("Online", { exact: true })).toBeVisible();
        await borrowerContext.setOffline(true);
        await expect(borrower.getByRole("status").filter({ hasText: "Offline" })).toBeVisible();
        await owner.getByLabel("Your message").fill("I'll bring the paperback edition.");
        await owner.getByRole("button", { name: "Send message", exact: true }).click();
        await expect(owner.getByRole("log")).toContainText("I'll bring the paperback edition.");
        await borrowerContext.setOffline(false);
        await expect(borrower.getByRole("status").filter({ hasText: "Connected" })).toBeVisible();
        await expect(borrower.getByRole("log")).toContainText("I'll bring the paperback edition.");
        await borrower.setViewportSize({ width: 390, height: 844 });
        expect(await borrower.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
        await borrower.screenshot({ path: testInfo.outputPath("messages-mobile.png"), fullPage: true });

        await visit(borrower, `/shelves/${account.user.id}`);
        await borrower.getByLabel("Message to the owner").fill("I'd love to borrow your copy.");
        await borrower.getByRole("button", { name: "Request to borrow", exact: true }).click();
        await expect(borrower.getByRole("status")).toContainText("Request sent");
        await visit(owner, "/requests");
        const loan = owner.locator("article").filter({ hasText: `Physical story ${stamp}` });
        await loan.getByRole("button", { name: "Approve", exact: true }).click();
        await expect(loan).toContainText("accepted");
        await loan.getByRole("button", { name: "Confirm handover", exact: true }).click();
        await expect(loan).toContainText("borrowed");
        await visit(borrower, "/requests");
        const borrowed = borrower.locator("article").filter({ hasText: `Physical story ${stamp}` });
        await borrowed.getByRole("link", { name: "Arrange exchange in messages" }).click();
        await expect(borrower.getByRole("log")).toContainText("I'd love to borrow your copy.");
        await borrower.getByLabel("Your message").fill("Thank you! I'll return it this afternoon.");
        await borrower.getByRole("button", { name: "Send message", exact: true }).click();
        await expect(borrower.getByRole("log")).toContainText("Thank you! I'll return it this afternoon.");
        await visit(borrower, "/requests");
        await borrowed.getByRole("button", { name: "I've returned the book", exact: true }).click();
        await expect(borrowed).toContainText("return pending");
        await expect(loan).toContainText("return pending");
        await loan.getByRole("button", { name: "Confirm receipt", exact: true }).click();
        await expect(loan).toContainText("returned");
        await visit(borrower, `/shelves/${account.user.id}`);
        await expect(borrower.getByRole("button", { name: "Request to borrow", exact: true })).toBeVisible();
        expect(errors).toEqual([]);
    } finally {
        if (copyId && ownerToken)
            await request.delete(`${booksApi}/copies/${copyId}`, {
                headers: { Authorization: `Bearer ${ownerToken}` },
            });
        await ownerContext.close();
        await borrowerContext.close();
    }
});
