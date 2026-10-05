import { env } from "$env/dynamic/private";
import { error, redirect, type RequestEvent } from "@sveltejs/kit";

const errorMessages: Record<string, string> = {
    UNAUTHORIZED: "A valid login is required",
    VALIDATION_ERROR: "Please check the highlighted fields.",
    INVALID_REQUEST_BODY: "Invalid input",
    BOOK_NOT_FOUND: "Book not found",
    AUTHOR_NOT_FOUND: "Author not found",
    INVALID_BOOK_ID: "Invalid book ID",
    AUTHOR_INVALID_ID: "Invalid author ID",
    INVALID_AUTHOR_ID: "Invalid author ID",
    INVALID_PUBLISHED_AFTER: "Enter a valid date.",
    INVALID_PUBLISHED_BEFORE: "Enter a valid date.",
    NO_FIELDS_TO_UPDATE: "At least one field must be provided to update.",
    BOOK_IN_USE: "This catalog title has physical copies and cannot be deleted",
    AUTHOR_IN_USE: "This author has catalog books and cannot be deleted",
};

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
        for (const entry of body?.errors ?? [])
            fields[entry.field] =
                entry.rule === "required"
                    ? "This field is required."
                    : entry.rule
                      ? "This field is invalid."
                      : entry.message;
        const message = Array.isArray(body?.message) ? body.message.join(". ") : body?.message;
        throw new ApiError(
            response.status,
            errorMessages[body?.code] ||
                (body?.code?.endsWith("_FAILED") ? "The request could not be completed." : message) ||
                "The request could not be completed.",
            fields
        );
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
