import { fail, redirect, type RequestEvent } from "@sveltejs/kit";
import { api, ApiError, requireUser } from "./api";
import type { Collection, FormValues } from "$lib/types";
import { readDate } from "$lib/i18n/calendar";

export async function save(event: RequestEvent, collection: Collection, id?: string) {
    requireUser(event);
    const form = await event.request.formData();
    const values: FormValues = {};
    for (const key of ["title", "author_id", "description", "published_at", "name", "bio"]) {
        values[key] = String(form.get(key) ?? "").trim();
    }
    const errors: FormValues = {};
    values.published_at = readDate(form, "published_at");
    const required = collection === "books" ? ["title", "author_id"] : ["name"];
    for (const key of required) if (!values[key]) errors[key] = "This field is required.";
    const textKey = collection === "books" ? "description" : "bio";
    if (values[textKey].length > 2000) errors[textKey] = "Use 2,000 characters or fewer.";
    if (collection === "books" && values.published_at && !/^\d{4}-\d{2}-\d{2}$/.test(values.published_at)) {
        errors.published_at = "Use a valid publication date.";
    }
    if (Object.keys(errors).length)
        return fail(400, { values, errors, message: "Please check the highlighted fields." });
    const body =
        collection === "books"
            ? {
                  title: values.title,
                  author_id: values.author_id,
                  description: values.description,
                  published_at: values.published_at,
              }
            : { name: values.name, bio: values.bio };
    let saved: { data: { id: string } };
    try {
        saved = await api(event, `/${collection}${id ? `/${encodeURIComponent(id)}` : ""}`, {
            method: id ? "PATCH" : "POST",
            body: JSON.stringify(body),
        });
    } catch (cause) {
        if (!(cause instanceof ApiError)) throw cause;
        if (cause.status === 401) redirect(303, `/login?returnTo=${encodeURIComponent(event.url.pathname)}`);
        return fail(cause.status, { values, errors: cause.fields, message: cause.message });
    }
    redirect(303, `/${collection}/${saved.data.id}`);
}
