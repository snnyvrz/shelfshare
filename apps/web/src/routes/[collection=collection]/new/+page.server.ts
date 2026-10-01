import { api, pageError, requireUser } from "$lib/server/api";
import { save } from "$lib/server/editor";
import type { Author, Collection } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async (event) => {
    requireUser(event);
    const collection = event.params.collection as Collection;
    try {
        const authors = collection === "books" ? await api<{ data: Author }[]>(event, "/authors") : [];
        return { collection, authors: authors.map((a) => a.data), values: {}, editing: false };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = { default: (event) => save(event, event.params.collection as Collection) };
