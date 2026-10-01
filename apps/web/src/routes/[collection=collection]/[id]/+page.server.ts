import { fail, redirect } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import type { Author, Book, Collection } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async (event) => {
    const collection = event.params.collection as Collection;
    try {
        const record = await api<{ data: Book | Author }>(
            event,
            `/${collection}/${encodeURIComponent(event.params.id)}`
        );
        return {
            collection,
            book: collection === "books" ? (record.data as Book) : null,
            author: collection === "authors" ? (record.data as Author) : null,
        };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = {
    delete: async (event) => {
        requireUser(event);
        try {
            await api(event, `/${event.params.collection}/${encodeURIComponent(event.params.id)}`, {
                method: "DELETE",
            });
        } catch (cause) {
            if (!(cause instanceof ApiError)) throw cause;
            if (cause.status === 401) redirect(303, `/login?returnTo=${encodeURIComponent(event.url.pathname)}`);
            return fail(cause.status, { message: cause.message });
        }
        redirect(303, `/${event.params.collection}`);
    },
};
