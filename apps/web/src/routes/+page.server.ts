import { api, ApiError } from "$lib/server/api";
import type { BookList } from "$lib/types";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async (event) => {
    try {
        const books = await api<BookList>(event, "/books?page_size=4");
        return { books: books.data, total: books.pagination.total, unavailable: false };
    } catch (cause) {
        if (!(cause instanceof ApiError)) throw cause;
        return { books: [], total: 0, unavailable: true };
    }
};
