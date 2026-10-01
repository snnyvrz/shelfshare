import { env } from "$env/dynamic/private";
import { error, redirect, type RequestEvent } from "@sveltejs/kit";

export class ApiError extends Error {
    constructor(
        public status: number,
        message: string,
        public fields: Record<string, string> = {}
    ) {
        super(message);
    }
}

export async function api<T>(event: RequestEvent, path: string, options: RequestInit = {}, auth = false): Promise<T> {
    const base = auth
        ? env.AUTH_API_URL || "http://localhost:3030/api/auth"
        : env.BOOKS_API_URL || "http://localhost:8080/api";
    const headers = new Headers(options.headers);
    headers.set("Accept", "application/json");
    if (options.body) headers.set("Content-Type", "application/json");
    if (event.locals.token) headers.set("Authorization", `Bearer ${event.locals.token}`);
    let response: Response;
    try {
        response = await event.fetch(`${base.replace(/\/$/, "")}${path}`, {
            ...options,
            headers,
            signal: AbortSignal.timeout(10000),
        });
    } catch {
        throw new ApiError(503, "The service is unavailable. Please try again shortly.");
    }
    if (response.status === 204) return undefined as T;
    const body = await response.json().catch(() => null);
    if (!response.ok) {
        const fields: Record<string, string> = {};
        for (const entry of body?.errors ?? []) fields[entry.field] = entry.message;
        const message = Array.isArray(body?.message) ? body.message.join(". ") : body?.message;
        throw new ApiError(response.status, message || "The request could not be completed.", fields);
    }
    return body as T;
}

export function requireUser(event: RequestEvent) {
    if (!event.locals.user)
        redirect(303, `/login?returnTo=${encodeURIComponent(event.url.pathname + event.url.search)}`);
}

export function pageError(cause: unknown): never {
    if (cause instanceof ApiError)
        error(cause.status >= 400 && cause.status <= 599 ? cause.status : 500, cause.message);
    throw cause;
}
