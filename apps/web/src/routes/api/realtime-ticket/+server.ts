import { env } from "$env/dynamic/private";
import { error, json } from "@sveltejs/kit";
import { api, ApiError } from "$lib/server/api";
import type { RequestHandler } from "./$types";
export const POST: RequestHandler = async (event) => {
    if (!event.locals.user) error(401, "Log in to connect");
    if (event.request.headers.get("origin") !== event.url.origin) error(403, "Invalid origin");
    try {
        const ticket = await api<{ ticket: string }>(event, "/ws-ticket", {
            method: "POST",
            body: JSON.stringify({ origin: event.url.origin }),
        });
        const url = new URL(env.BOOKS_WS_URL || `${env.BOOKS_API_URL || "http://localhost:8080/api"}/ws`);
        url.protocol = url.protocol === "https:" || url.protocol === "wss:" ? "wss:" : "ws:";
        url.searchParams.set("ticket", ticket.ticket);
        return json({ url: url.toString() }, { headers: { "Cache-Control": "no-store" } });
    } catch (cause) {
        if (!(cause instanceof ApiError)) throw cause;
        return json({ message: cause.message }, { status: cause.status });
    }
};
