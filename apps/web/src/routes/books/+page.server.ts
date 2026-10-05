import { api, pageError } from "$lib/server/api";
import type { Author, BookList } from "$lib/types";
import type { PageServerLoad } from "./$types";
import { readDate } from "$lib/i18n/calendar";

export const load: PageServerLoad = async (event) => {
    const query = new URLSearchParams({ page_size: "12" });
    for (const key of ["q", "sort", "author_id", "published_after", "published_before", "page"]) {
        const value = key.startsWith("published_")
            ? readDate(event.url.searchParams, key)
            : event.url.searchParams.get(key);
        if (value) query.set(key, value);
    }
    try {
        const [books, authors] = await Promise.all([
            api<BookList>(event, `/books?${query}`),
            api<{ data: Author }[]>(event, "/authors"),
        ]);
        return { books, authors: authors.map((a) => a.data), filters: Object.fromEntries(query) };
    } catch (cause) {
        pageError(cause);
    }
};
