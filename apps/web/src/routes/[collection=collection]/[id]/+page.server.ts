import { fail, redirect } from "@sveltejs/kit";
import { api, ApiError, pageError, requireUser } from "$lib/server/api";
import type { Author, Book, Collection } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";
import { borrow, copyList } from "$lib/server/community";

export const load: PageServerLoad = async (event) => {
    event.depends("shelfshare:community");
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
            ...(collection === "books"
                ? await copyList(event, `/copies?bookId=${encodeURIComponent(event.params.id)}`)
                : { copies: [], profiles: {}, total: 0 }),
        };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = {
    borrow,
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
