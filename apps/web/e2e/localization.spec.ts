import { test, expect, type BrowserContext, type Page } from "@playwright/test";
import { SignJWT } from "jose";

// Run with playwright.i18n.config.ts; all APIs are disposable in-memory fixtures.
async function visit(page: Page, path: string) {
    await page.goto(path);
    await expect(page.locator("html")).toHaveAttribute("data-hydrated", "true");
}
async function signIn(context: BrowserContext) {
    const token = await new SignJWT({ email: "reader@example.com" })
        .setProtectedHeader({ alg: "HS256" })
        .setSubject("000000000000000000000001")
        .setExpirationTime("1h")
        .sign(new TextEncoder().encode("localization-test-secret"));
    await context.addCookies([
        { name: "shelfshare_session", value: token, url: "http://localhost:5174", httpOnly: true },
    ]);
}

test("concurrent SSR requests keep locale preferences isolated and reject invalid switch input", async ({
    request,
}) => {
    const [fa, en] = await Promise.all([
        request.get("/login", { headers: { Cookie: "shelfshare_locale=invalid" } }),
        request.get("/login", { headers: { Cookie: "shelfshare_locale=en" } }),
    ]);
    expect(await fa.text()).toContain('lang="fa" dir="rtl"');
    expect(await en.text()).toContain('lang="en" dir="ltr"');
    const invalid = await request.post("/api/locale", { form: { locale: "fr" } });
    expect(invalid.status()).toBe(400);
    const switchResult = await request.post("/api/locale", {
        form: { locale: "fa", returnTo: "//example.com" },
        maxRedirects: 0,
    });
    expect(switchResult.status()).toBe(303);
    expect(switchResult.headers().location).toBe("/");
});

test("Persian SSR is default; language switching preserves filters and form drafts", async ({ page, request }) => {
    const html = await (await request.get("/login")).text();
    expect(html).toContain('lang="fa" dir="rtl"');
    expect(html).toContain("نشانی ایمیل");
    await visit(page, "/register?returnTo=%2Fmy-shelf");
    await page.getByLabel("نام عمومی").fill("مریم / Maryam");
    await page.getByRole("button", { name: "زبان", exact: true }).click();
    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    await expect(page.locator("html")).toHaveAttribute("dir", "ltr");
    await expect(page.getByLabel("Public display name")).toHaveValue("مریم / Maryam");
    await expect(page).toHaveURL(/returnTo=%2Fmy-shelf/);
    await page.reload();
    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    await page.getByRole("button", { name: "Language", exact: true }).click();
    await expect(page.locator("html")).toHaveAttribute("dir", "rtl");
});

test("cities, Jalali dates, mobile layout, and editor/composer drafts follow the language", async ({
    page,
    context,
}) => {
    await signIn(context);
    await visit(page, "/account/profile");
    await expect(page.locator('select[name="discoveryCityId"] option:checked')).toHaveText("تهران، ایران");
    await page.getByRole("button", { name: "زبان", exact: true }).click();
    await expect(page.locator('select[name="discoveryCityId"] option:checked')).toHaveText("Tehran, Iran");
    await visit(page, "/shelves");
    await expect(page.getByText("Tehran, Iran", { exact: true }).last()).toBeVisible();
    await page.getByRole("button", { name: "Language", exact: true }).click();
    await expect(page.getByText("تهران، ایران", { exact: true }).last()).toBeVisible();
    await page.setViewportSize({ width: 390, height: 844 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await visit(page, "/books/new");
    await page.getByLabel("عنوان کتاب").fill("پیش‌نویس / Draft");
    await page.getByLabel("تاریخ انتشار").fill("۱۴۰۵/۰۷/۱۳");
    await page.getByRole("button", { name: "زبان", exact: true }).click();
    await expect(page.getByLabel("Book title")).toHaveValue("پیش‌نویس / Draft");
    await expect(page.getByLabel("Publication date")).toHaveValue("2026-10-05");
    await page.getByRole("button", { name: "Language", exact: true }).click();
    await expect(page.getByLabel("تاریخ انتشار")).toHaveValue("۱۴۰۵/۰۷/۱۳");
    await page.getByLabel("نویسنده *", { exact: true }).selectOption("author");
    await page.getByRole("button", { name: "افزودن کتاب", exact: true }).click();
    await expect(page).toHaveURL(/\/books\/book$/);
    await expect(page.getByText(/منتشرشده در.*۱۴۰۵/)).toBeVisible();
    await visit(page, "/messages?conversation=conversation");
    await page.getByLabel("پیام شما").fill("سلام / Hello draft");
    await page.getByRole("button", { name: "زبان", exact: true }).click();
    await expect(page.getByLabel("Your message")).toHaveValue("سلام / Hello draft");
});

test("Persian errors and due-date conversion reach the API", async ({ page, context, request }) => {
    await visit(page, "/login");
    await page.getByLabel("نشانی ایمیل").fill("reader@example.com");
    await page.getByLabel("گذرواژه", { exact: true }).fill("wrong-password");
    await page.getByRole("button", { name: "ورود", exact: true }).click();
    await expect(page.getByRole("alert")).toContainText("ایمیل یا گذرواژه نادرست است");
    await signIn(context);
    await visit(page, "/requests");
    await page.getByLabel("تاریخ توافق‌شدهٔ بازگشت (اختیاری)").fill("۱۴۰۶/۰۱/۰۱");
    await page.getByRole("button", { name: "تأیید تحویل کتاب", exact: true }).click();
    await expect(page.getByRole("status")).toContainText("درخواست به‌روز شد");
    const state = await (await request.get("http://127.0.0.1:31303/test/state")).json();
    expect(state.lastDueAt).toBe("2027-03-21T23:59:59.000Z");
});

test("non-JavaScript forms switch language and submit Jalali dates", async ({ browser, request }) => {
    const context = await browser.newContext({ javaScriptEnabled: false });
    try {
        await signIn(context);
        const page = await context.newPage();
        await page.goto("http://localhost:5174/books/new");
        await page.getByLabel("عنوان کتاب").fill("کتاب بدون جاوااسکریپت");
        await page.getByLabel("نویسنده *", { exact: true }).selectOption("author");
        await page.getByLabel("تاریخ انتشار").fill("۱۳۹۹/۱۲/۳۰");
        await page.getByRole("button", { name: "افزودن کتاب", exact: true }).click();
        await expect(page).toHaveURL(/\/books\/book$/);
        const state = await (await request.get("http://127.0.0.1:31303/test/state")).json();
        expect(state.book.published_at).toBe("2021-03-20");
        await page.getByRole("button", { name: "زبان", exact: true }).click();
        await expect(page.locator("html")).toHaveAttribute("dir", "ltr");
    } finally {
        await context.close();
    }
});
