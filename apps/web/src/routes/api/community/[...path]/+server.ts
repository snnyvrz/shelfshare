import { error, json } from "@sveltejs/kit";
import { api, ApiError } from "$lib/server/api";
import type { RequestHandler } from "./$types";

const proxy: RequestHandler = async (event) => {
    if (!event.locals.user) error(401, "Log in to continue");
    if (event.request.method !== "GET" && event.request.headers.get("origin") !== event.url.origin)
        error(403, "Invalid origin");
    const path = event.params.path;
    if (!/^(conversations|blocks)(\/|$)/.test(path) || path.includes("..")) error(404);
    try {
        const body =
            event.request.method === "GET" || event.request.method === "DELETE"
                ? undefined
                : await event.request.text();
        const result = await api(event, `/${path}${event.url.search}`, {
            method: event.request.method,
            ...(body ? { body } : {}),
        });
        return result === undefined ? new Response(null, { status: 204 }) : json(result);
    } catch (cause) {
        if (!(cause instanceof ApiError)) throw cause;
        return json({ message: cause.message }, { status: cause.status });
    }
};
export const GET = proxy;
export const POST = proxy;
export const DELETE = proxy;
