import { env } from "$env/dynamic/private";
import { fail, redirect } from "@sveltejs/kit";
import { api, ApiError } from "$lib/server/api";
import { safeReturnTo, sessionCookie, verifySession } from "$lib/server/session";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = ({ params, url, locals }) => {
    const returnTo = safeReturnTo(url.searchParams.get("returnTo"));
    if (locals.user) redirect(303, returnTo);
    return { mode: params.mode, returnTo };
};

export const actions: Actions = {
    default: async (event) => {
        const form = await event.request.formData();
        const email = String(form.get("email") ?? "")
            .trim()
            .toLowerCase();
        const password = String(form.get("password") ?? "");
        const returnTo = safeReturnTo(event.url.searchParams.get("returnTo"));
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) || !password) {
            return fail(400, { email, message: "Enter a valid email address and password." });
        }
        let result: { token: string };
        try {
            result = await api(
                event,
                `/${event.params.mode}`,
                {
                    method: "POST",
                    body: JSON.stringify({
                        email,
                        password,
                        ...(event.params.mode === "register"
                            ? { displayName: String(form.get("displayName") || "Reader") }
                            : {}),
                    }),
                },
                true
            );
        } catch (cause) {
            if (!(cause instanceof ApiError)) throw cause;
            return fail(cause.status, { email, message: cause.message });
        }
        const session = await verifySession(result.token, env.JWT_SECRET ?? "");
        if (!session)
            return fail(502, { email, message: "Unable to establish a session. Check the shared JWT configuration." });
        event.cookies.set(sessionCookie, result.token, {
            path: "/",
            httpOnly: true,
            sameSite: "lax",
            secure: event.url.protocol === "https:",
            expires: session.expires,
        });
        redirect(303, returnTo);
    },
};
