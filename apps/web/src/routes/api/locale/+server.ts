import { json, redirect } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
import { localeCookie } from "$lib/i18n";

export const POST: RequestHandler = async ({ request, cookies, url }) => {
    const form = await request.formData();
    const locale = form.get("locale");
    if (locale !== "fa" && locale !== "en") return json({ error: "Invalid locale" }, { status: 400 });
    cookies.set(localeCookie, locale, {
        path: "/",
        httpOnly: true,
        sameSite: "lax",
        secure: url.protocol === "https:",
        maxAge: 60 * 60 * 24 * 365,
    });
    if (request.headers.get("accept") === "application/json") return json({ locale });
    const returnTo = form.get("returnTo");
    const destination =
        typeof returnTo === "string" &&
        returnTo.startsWith("/") &&
        !returnTo.startsWith("//") &&
        !returnTo.includes("\\")
            ? returnTo
            : "/";
    redirect(303, destination);
};
