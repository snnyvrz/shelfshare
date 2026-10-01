import { env } from "$env/dynamic/private";
import type { Handle } from "@sveltejs/kit";
import { sessionCookie, verifySession } from "$lib/server/session";

export const handle: Handle = async ({ event, resolve }) => {
    event.locals.user = null;
    event.locals.token = null;
    const token = event.cookies.get(sessionCookie);
    if (token) {
        const session = await verifySession(token, env.JWT_SECRET ?? "");
        if (session) {
            event.locals.user = session.user;
            event.locals.token = token;
        } else {
            event.cookies.delete(sessionCookie, { path: "/" });
        }
    }
    return resolve(event);
};
