import { api, pageError, requireUser } from "$lib/server/api";
import { save } from "$lib/server/editor";
import type { Author, Book, Collection, FormValues } from "$lib/types";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async (event) => {
    requireUser(event);
    const collection = event.params.collection as Collection;
    try {
        const record = await api<{ data: Book | Author }>(
            event,
            `/${collection}/${encodeURIComponent(event.params.id)}`
        );
        const authors = collection === "books" ? await api<{ data: Author }[]>(event, "/authors") : [];
        const values: FormValues =
            collection === "books"
                ? {
                      title: (record.data as Book).title,
                      author_id: (record.data as Book).author.id,
                      description: (record.data as Book).description,
                      published_at: (record.data as Book).published_at ?? "",
                  }
                : { name: (record.data as Author).name, bio: (record.data as Author).bio };
        return { collection, authors: authors.map((a) => a.data), values, editing: true };
    } catch (cause) {
        pageError(cause);
    }
};
export const actions: Actions = {
    default: (event) => save(event, event.params.collection as Collection, event.params.id),
};
